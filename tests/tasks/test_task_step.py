import json

import pytest

from tests.conftest import CONFIG
from tests.tasks.fake_runner import fake_runner
from worker.config.load_worker_config import load_worker_config
from worker.node.host_state import HostState
from worker.tasks.task_step import task_step


class Link:
    def __init__(self, lease):
        self.lease, self.leases, self.completed = lease, [], []

    def lease_task(self, user_active, idle_s):
        self.leases.append(user_active)
        return self.lease

    def heartbeat(self, attempt, generation, draining):
        return True

    def complete(self, attempt, generation, report):
        self.completed.append((attempt, generation, report))


def make_config(tmp_path, command):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"p": {"runner": "codex", "command": command}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def step(config, link, state):
    return task_step(config, "node-a", link, lambda: state, lambda: 0.0, lambda _s: None)


LEASE = {"attempt_id": "a1", "generation": 2, "model": "p", "input": {"prompt": "x"}}
ANSWER = 'open(args[args.index("-o") + 1], "w").write("{}")'


def test_a_task_is_leased_run_and_completed(tmp_path):
    link = Link(LEASE)
    config = make_config(tmp_path, fake_runner(tmp_path, "codex", ANSWER))
    assert step(config, link, HostState(5.0, True, "warn")) is False
    assert link.leases == [True]  # idle 5 s is below the threshold: the user is active
    attempt, generation, report = link.completed[0]
    assert (attempt, generation, report["outcome"]) == ("a1", 2, "succeeded")


@pytest.mark.parametrize(
    "state", [HostState(900.0, False, "normal"), HostState(None, True, "normal")]
)
def test_battery_or_an_unreadable_host_takes_no_task(tmp_path, state):
    link = Link(LEASE)
    assert step(make_config(tmp_path, "codex"), link, state) is True
    assert link.leases == []


def test_a_lease_for_a_removed_profile_fails_as_unknown_profile(tmp_path):
    link = Link({**LEASE, "model": "gone"})
    step(make_config(tmp_path, "codex"), link, HostState(900.0, True, "normal"))
    assert link.completed[0][2]["error_code"] == "unknown_profile"


def test_agy_profiles_take_no_input_root(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"p": {"runner": "agy", "input_root": "x"}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    with pytest.raises(ValueError, match="agy"):
        load_worker_config(path)
