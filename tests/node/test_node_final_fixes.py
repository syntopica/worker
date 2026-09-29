import dataclasses
import time

import pytest

from tests.node.fake_ollama import FakeOllama
from tests.node.test_run_node_fixes import LEASE, PIN, Link, StopLoopError, node
from worker.node import relieve_pressure as relieve_pressure_module
from worker.node import run_leased as run_leased_module
from worker.node.drain_backend import drain_backend
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory
from worker.node.run_attempt import run_attempt
from worker.node.run_node import run_node
from worker.node.settle_attempt import settle_attempt

IDLE = HostState(900, True, "normal")
WARN = HostState(900, True, "warn")


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def loop(config, url, link, rests, sample=lambda: IDLE):
    local = dataclasses.replace(config.nodes["node-a"], ollama_url=url)
    config = dataclasses.replace(config, nodes={**config.nodes, "node-a": local})
    seen = []
    ticks = iter([i * 200.0 for i in range(200)])

    def sleep(seconds):
        seen.append(seconds)
        if len(seen) >= rests:
            raise StopLoopError

    with pytest.raises(StopLoopError):
        run_node(config, "node-a", link, sample=sample, sleep=sleep, clock=lambda: next(ticks))


def test_an_attempt_that_raises_is_reported_as_node_error(config, fake, monkeypatch, capsys):
    def explode(*_args):
        raise RuntimeError("generated text must never be logged")

    monkeypatch.setattr(run_leased_module, "run_attempt", explode)
    link = Link([LEASE])
    loop(config, fake.url, link, rests=3)
    assert [(r["outcome"], r["error_code"]) for r in link.completed] == [("failed", "node_error")]
    err = capsys.readouterr().err
    assert "RuntimeError" in err
    assert "generated text" not in err
    assert len(link.reasons) >= 3  # the loop kept going


def test_a_failing_iteration_is_logged_and_the_loop_continues(config, fake, capsys):
    class Flaky(Link):
        def report(self, state):
            if not self.reasons:
                self.reasons.append("boom")
                raise ValueError("secret")
            super().report(state)

    link = Flaky()
    loop(config, fake.url, link, rests=3)
    assert link.reasons[:2] == ["boom", None]
    err = capsys.readouterr().err
    assert "worker: node iteration failed: ValueError" in err
    assert "secret" not in err


def test_an_unreachable_backend_blocks_admission(config):
    class NoLease(Link):
        def lease(self, *args):
            raise AssertionError("must not lease while the backend is down")

    link = NoLease()
    loop(config, "http://127.0.0.1:1", link, rests=3)
    assert link.reasons[1:] == ["backend_down", "backend_down"]


def test_a_transport_error_charges_one_attempt_per_outage(config, fake, monkeypatch):
    def transport_error(*args):
        fake.ps_raw = b"not json"  # the backend went away with the attempt
        args[-1]("transport_error")
        return "failed"

    monkeypatch.setattr(run_leased_module, "run_attempt", transport_error)
    link = Link([LEASE, LEASE])
    loop(config, fake.url, link, rests=4)
    assert link.leases == [LEASE]
    assert link.reasons[2:] == ["backend_down", "backend_down"]


def test_pressure_releases_only_after_three_consecutive_samples(fake):
    fake.delay = 3.0
    seen = []

    def sample():
        seen.append(1)
        return WARN

    outcome = run_attempt(LEASE, Link(), PIN, node(fake.url), sample, lambda s: None, time.time)
    assert (outcome, len(seen)) == ("preempted", 3)


def test_intermittent_pressure_does_not_release(fake):
    fake.delay = 0.6
    pattern = iter([WARN, WARN, IDLE] * 20)
    link = Link()
    outcome = run_attempt(
        LEASE,
        link,
        PIN,
        node(fake.url),
        lambda: next(pattern),
        lambda s: time.sleep(0.05),
        time.time,
    )
    assert outcome == "succeeded"


def test_only_a_model_the_node_loaded_is_unloaded_on_pressure(config, fake, monkeypatch):
    fake.loaded = []
    unloaded = []
    monkeypatch.setattr(
        relieve_pressure_module, "unload_model", lambda url, model: unloaded.append(model) or True
    )

    def loads_the_model(*_args):
        fake.loaded = ["model-a"]
        return "succeeded"

    monkeypatch.setattr(run_leased_module, "run_attempt", loads_the_model)
    samples = iter([IDLE, IDLE] + [HostState(900, True, "critical")] * 20)
    loop(config, fake.url, Link([LEASE]), rests=7, sample=lambda: next(samples))
    assert unloaded == ["model-a"]  # after three samples, once per episode


def test_a_pressure_release_is_followed_by_a_reported_backoff(config, fake, monkeypatch):
    def released(*args):
        args[-1]("memory_pressure")
        return "preempted"

    monkeypatch.setattr(run_leased_module, "run_attempt", released)
    link = Link([LEASE])
    loop(config, fake.url, link, rests=4)
    assert link.reasons[2:4] == ["pressure_backoff", "pressure_backoff"]


def test_the_backoff_doubles_caps_at_two_hours_and_resets_on_success():
    memory = NodeMemory()
    ends = []
    for _ in range(5):
        settle_attempt(memory, "model-a", True, "preempted", "memory_pressure", 0.0)
        ends.append(memory.backoff_until)
    assert ends == [900.0, 1800.0, 3600.0, 7200.0, 7200.0]
    settle_attempt(memory, "model-a", True, "succeeded", None, 10.0)
    settle_attempt(memory, "model-a", True, "preempted", "memory_pressure", 10.0)
    assert memory.backoff_until == 910.0


def test_a_model_that_is_not_resident_is_quiet_without_a_probe():
    probes = []
    assert drain_backend(
        "u",
        PIN,
        "label",
        lambda: True,
        probe=lambda *a, **k: probes.append(1),
        resident=lambda _u: [],
    )
    assert probes == []


def test_after_an_unload_quiet_comes_from_ps_and_no_restart_follows():
    steps = []
    listings = iter([["model-a"]] * 3 + [[]])
    assert drain_backend(
        "u",
        PIN,
        "label",
        lambda: steps.append("beat") or True,
        probe=lambda *a, **k: steps.append("probe") and False,
        unload=lambda *a: steps.append("unload") or True,
        restart=lambda *a: steps.append("restart") or True,
        resident=lambda _u: next(listings),
        pause=lambda _s: None,
    )
    assert steps == ["beat", "probe", "beat", "probe", "beat", "probe", "beat", "unload", "beat"]


def test_an_unreachable_ps_after_unload_never_restarts():
    calls = []
    listings = iter([["model-a"]] * 3 + [None] * 10)
    assert not drain_backend(
        "u",
        PIN,
        "label",
        lambda: True,
        probe=lambda *a, **k: False,
        unload=lambda *a: calls.append("unload") or True,
        restart=lambda *a: calls.append("restart") or True,
        resident=lambda _u: next(listings),
        pause=lambda _s: None,
    )
    assert calls == ["unload"]
