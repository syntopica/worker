import threading

import pytest

from tests.remote.test_remote_step import LEASE, Link
from worker.remote import run_remote_attempt as attempt_module
from worker.remote.run_remote_attempt import run_remote_attempt
from worker.remote.run_remote_slots import run_remote_slots


def test_every_slot_runs_and_the_others_stop_with_the_first(config):
    seen = []
    ended = []

    def run(cfg, node, link, key, sleep=None, *, current=None, stopping=None):
        seen.append(link)
        if stopping is None:
            raise SystemExit(0)
        while not stopping():
            sleep(0.01)
        ended.append(link)

    with pytest.raises(SystemExit):
        run_remote_slots(config, "node-a", ["l0", "l1", "l2"], "k", run=run)
    assert sorted(seen) == ["l0", "l1", "l2"]
    assert sorted(ended) == ["l1", "l2"]


def test_a_stopping_slot_hands_its_running_call_back(config, monkeypatch):
    monkeypatch.setattr(attempt_module, "_BEAT_S", 0.01)
    release = threading.Event()
    link = Link()

    def post(key, body, timeout):
        release.wait(5)
        return None, "transport_error"

    try:
        outcome = run_remote_attempt(
            LEASE, link, "node-a", "k", config, lambda: 0.0, post, stopping=lambda: True
        )
    finally:
        release.set()
    assert outcome == "preempted"
    assert link.completed[0]["error_code"] == "node_shutdown"
