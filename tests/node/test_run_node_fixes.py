import dataclasses
import time

import pytest

from tests.node.fake_ollama import FakeOllama
from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.node import relieve_pressure as relieve_pressure_module
from worker.node import run_attempt as run_attempt_module
from worker.node import run_leased as run_leased_module
from worker.node.drain_backend import drain_backend
from worker.node.host_state import HostState
from worker.node.run_attempt import run_attempt
from worker.node.run_node import run_node

PIN = ModelPin("model-a", 40960, "5m", 30, 2)
IDLE = HostState(900, True, "normal")
LEASE = {
    "job_id": "j",
    "attempt_id": "a",
    "generation": 1,
    "model": "model-a",
    "input": {"messages": []},
    "run_when": "idle",
}


class StopLoopError(Exception):
    pass


class Link:
    def __init__(self, leases=(), alive=True):
        self.leases = list(leases)
        self.alive = alive
        self.reasons = []
        self.completed = []

    def heartbeat(self, attempt_id, generation, draining):
        return self.alive

    def complete(self, attempt_id, generation, report):
        self.completed.append(report)

    def report(self, state):
        self.reasons.append(state["reason"])

    def lease(self, *args):
        return self.leases.pop(0) if self.leases else None


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def node(url):
    return NodePolicy("node-a", "owner", 300, 44, url, "label-a")


def drive(config, fake, link, rests, *, state=IDLE, times=None):
    """Run the loop forever against the fake backend; stop on the n-th rest."""
    local = dataclasses.replace(config.nodes["node-a"], ollama_url=fake.url)
    config = dataclasses.replace(config, nodes={**config.nodes, "node-a": local})
    seen = []
    ticks = iter(times or [i * 200.0 for i in range(50)])

    def sleep(seconds):
        seen.append(seconds)
        if len(seen) >= rests:
            raise StopLoopError

    with pytest.raises(StopLoopError):
        run_node(
            config, "node-a", link, sample=lambda: state, sleep=sleep, clock=lambda: next(ticks)
        )


def test_drain_failed_node_readmits_once_ps_drops_the_model_without_probing(
    config, fake, monkeypatch
):
    fake.loaded = []

    def failed_drain(*_args):
        fake.loaded = ["model-a"]  # the model stayed resident after the failed drain
        return "drain_failed"

    monkeypatch.setattr(run_leased_module, "run_attempt", failed_drain)
    link = Link([LEASE])
    rests = []

    def sleep(seconds):
        rests.append(seconds)
        if len(rests) == 3:
            fake.loaded = []  # keep-alive expired: /api/ps no longer lists it
        if len(rests) >= 4:
            raise StopLoopError

    local = dataclasses.replace(config.nodes["node-a"], ollama_url=fake.url)
    config = dataclasses.replace(config, nodes={**config.nodes, "node-a": local})
    ticks = iter([i * 200.0 for i in range(50)])
    with pytest.raises(StopLoopError):
        run_node(
            config, "node-a", link, sample=lambda: IDLE, sleep=sleep, clock=lambda: next(ticks)
        )
    assert link.reasons == ["pressure_recovering", None, "drain_failed", "drain_failed", None]
    assert not [b for b in fake.bodies if b.get("options", {}).get("num_predict") == 1]


def test_fenced_path_reports_a_failed_drain(fake, monkeypatch):
    fake.delay = 3.0
    monkeypatch.setattr(run_attempt_module, "drain_backend", lambda *a: False)
    outcome = run_attempt(
        LEASE,
        Link(alive=False),
        PIN,
        node(fake.url),
        lambda: IDLE,
        lambda s: time.sleep(0.05),
        time.time,
    )
    assert outcome == "drain_failed"


def test_a_result_that_lands_during_sampling_is_not_preempted(fake):
    fake.delay = 0.3
    link = Link()

    def slow_typing():
        time.sleep(0.5)
        return HostState(1, True, "normal")

    outcome = run_attempt(
        LEASE, link, PIN, node(fake.url), slow_typing, lambda s: time.sleep(0.05), time.time
    )
    assert (outcome, link.completed[0]["outcome"]) == ("succeeded", "succeeded")


def test_fenced_drain_never_unloads_or_restarts():
    calls = []
    quiet = drain_backend(
        "u",
        PIN,
        "label",
        lambda: False,
        probe=lambda *a, **k: False,
        resident=lambda _u: ["model-a"],
        unload=lambda *a: calls.append("unload") or True,
        restart=lambda *a: calls.append("restart") or True,
    )
    assert (quiet, calls) == (False, [])


def test_fenced_drain_still_reports_a_quiet_backend():
    assert drain_backend(
        "u", PIN, "label", lambda: False, probe=lambda *a, **k: True, resident=lambda _u: [PIN.name]
    )


def test_a_clock_starting_at_zero_still_admits_work(config, fake):
    link = Link()
    drive(config, fake, link, rests=2, times=[0.0, 200.0, 400.0])
    assert link.reasons == ["pressure_recovering", None]


def test_unknown_model_is_failed_and_the_loop_continues(config, fake):
    link = Link([{**LEASE, "model": "model-zz"}])
    drive(config, fake, link, rests=2)
    assert [(r["outcome"], r["error_code"]) for r in link.completed] == [
        ("failed", "unknown_model")
    ]


def test_sustained_pressure_never_unloads_a_model_resident_at_startup(config, fake, monkeypatch):
    unloaded = []
    monkeypatch.setattr(
        relieve_pressure_module, "unload_model", lambda url, model: unloaded.append(model) or True
    )
    link = Link()
    drive(config, fake, link, rests=5, state=HostState(900, True, "critical"))
    assert (link.reasons[-1], unloaded) == ("memory_pressure", [])


def test_the_startup_recovery_window_does_not_unload_the_model(config, fake, monkeypatch):
    unloaded = []
    monkeypatch.setattr(
        relieve_pressure_module, "unload_model", lambda url, model: unloaded.append(model) or True
    )
    link = Link()
    drive(config, fake, link, rests=1)
    assert (link.reasons, unloaded) == (["pressure_recovering"], [])
