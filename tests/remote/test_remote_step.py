import time

from tests.remote.fake_openrouter import FakeOpenRouter
from worker.remote import run_remote_attempt as attempt_module
from worker.remote.remote_step import remote_step
from worker.remote.run_remote_attempt import run_remote_attempt

LEASE = {
    "job_id": "j",
    "attempt_id": "a",
    "generation": 1,
    "model": "vendor/model:free",
    "input": {"messages": [{"role": "user", "content": "hi"}]},
    "queue": "pa.bulk",
    "privacy": "internal",
}


class Link:
    def __init__(self, lease=None, alive=True):
        self.lease = lease
        self.alive = alive
        self.completed = []

    def lease_remote(self):
        return self.lease

    def heartbeat(self, attempt_id, generation, draining):
        return self.alive

    def complete(self, attempt_id, generation, report):
        self.completed.append(report)


def test_no_headroom_or_unknown_headroom_rests_without_leasing(config):
    link = Link(LEASE)
    assert remote_step(config, "node-a", link, "k", lambda: 0.0, headroom=lambda _k: 0) == 900.0
    assert remote_step(config, "node-a", link, "k", lambda: 0.0, headroom=lambda _k: None) == 900.0
    assert link.completed == []


def test_nothing_to_escalate_rests_briefly(config):
    assert remote_step(config, "node-a", Link(), "k", lambda: 0.0, headroom=lambda _k: 3) == 30.0


def test_a_run_reports_the_remote_executor_and_zdr_for_non_public(config, monkeypatch):
    fake = FakeOpenRouter()
    try:
        monkeypatch.setattr(attempt_module, "_BEAT_S", 0.01)
        link = Link()

        def post(key, body, timeout):
            return attempt_module.post_openrouter(key, body, timeout, api=fake.api)

        outcome = run_remote_attempt(LEASE, link, "node-a", "k", config, lambda: 0.0, post)
    finally:
        fake.close()
    assert outcome == "succeeded"
    assert link.completed[0]["executor"]["provider"] == "openrouter"
    assert link.completed[0]["output"]["json"] == {"label": "x"}
    assert fake.bodies[0]["provider"] == {"zdr": True}


def test_a_fenced_remote_attempt_is_abandoned(config, monkeypatch):
    monkeypatch.setattr(attempt_module, "_BEAT_S", 0.01)
    link = Link(alive=False)

    def slow(key, body, timeout):
        time.sleep(0.2)
        return None, "transport_error"

    assert run_remote_attempt(LEASE, link, "node-a", "k", config, lambda: 0.0, slow) == "fenced"
    assert link.completed == []
