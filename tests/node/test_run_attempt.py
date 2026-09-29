import time

import pytest

from tests.node.fake_ollama import FakeOllama
from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.node.drain_backend import drain_backend
from worker.node.host_state import HostState
from worker.node.run_attempt import run_attempt
from worker.node.run_node import run_node

PIN = ModelPin("model-a", 40960, "5m", 30, 2)
IDLE = HostState(900, True, "normal")
TYPING = HostState(1, True, "normal")


class Link:
    def __init__(self, alive=True):
        self.alive = alive
        self.completed = []
        self.beats = []

    def heartbeat(self, attempt_id, generation, draining):
        self.beats.append(draining)
        return self.alive

    def complete(self, attempt_id, generation, report):
        self.completed.append(report)


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def lease(run_when="idle"):
    return {
        "job_id": "j",
        "attempt_id": "a",
        "generation": 1,
        "model": "model-a",
        "input": {"messages": []},
        "run_when": run_when,
    }


def node(url):
    return NodePolicy("node-a", "owner", 300, 44, url, "label-a")


def states(*items):
    seq = list(items)
    return lambda: seq.pop(0) if len(seq) > 1 else seq[0]


def nap(_seconds):
    time.sleep(0.05)


def test_success_is_reported_with_output_and_wall_time(fake):
    link = Link()
    clock = iter(range(100)).__next__
    outcome = run_attempt(lease(), link, PIN, node(fake.url), states(IDLE), lambda s: None, clock)
    assert outcome == "succeeded"
    assert link.completed[0]["output"]["json"] == {"label": "x"}
    assert link.completed[0]["wall_s"] > 0


def test_backend_error_is_reported_as_failed(fake):
    fake.status = 500
    fake.raw_body = b"{}"
    link = Link()
    outcome = run_attempt(lease(), link, PIN, node(fake.url), states(IDLE), nap, time.time)
    assert outcome == "failed"
    assert link.completed[0]["error_code"] == "http_500"


def test_user_return_preempts_after_draining(fake):
    fake.delay = 3.0
    link = Link()
    outcome = run_attempt(lease(), link, PIN, node(fake.url), states(IDLE, TYPING), nap, time.time)
    assert outcome == "preempted"
    assert link.completed[0]["outcome"] == "preempted"
    assert link.completed[0]["error_code"] == "user_active"
    assert True in link.beats


def test_active_ok_keeps_running_while_the_user_types(fake):
    fake.delay = 0.3
    link = Link()
    outcome = run_attempt(
        lease("active_ok"), link, PIN, node(fake.url), states(TYPING), nap, time.time
    )
    assert outcome == "succeeded"


def test_fenced_attempt_stops_without_completing(fake):
    fake.delay = 3.0
    link = Link(alive=False)
    outcome = run_attempt(lease(), link, PIN, node(fake.url), states(IDLE), nap, time.time)
    assert (outcome, link.completed) == ("fenced", [])


def test_drain_escalates_to_unload_then_restart():
    calls = []
    probes = iter([False] * 6 + [True])
    assert drain_backend(
        "u",
        PIN,
        "label",
        lambda: True,
        probe=lambda *a, **k: next(probes),
        unload=lambda *a: calls.append("unload") or True,
        restart=lambda *a: calls.append("restart") or True,
    )
    assert calls == ["unload", "restart"]


def test_drain_failure_is_reported_when_nothing_answers():
    assert not drain_backend(
        "u",
        PIN,
        "label",
        lambda: True,
        probe=lambda *a, **k: False,
        unload=lambda *a: True,
        restart=lambda *a: True,
    )


def test_node_reports_why_it_is_not_working(config):
    reports = []

    class L(Link):
        def report(self, state):
            reports.append(state)

        def lease(self, *args):
            raise AssertionError("must not lease while on battery")

    run_node(
        config,
        "node-a",
        L(),
        sample=lambda: HostState(900, False, "normal"),
        sleep=lambda s: None,
        clock=lambda: 1000.0,
        forever=False,
    )
    assert reports[0]["reason"] == "on_battery"
