import json

import pytest

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job

MAX = {"runner": "max-lane", "model": "claude-x", "on_demand": True, "privacy": ["public"]}


def make_config(tmp_path, profile=MAX, route=False):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"max": profile}
    if route:
        raw["queues"]["pa.bulk"]["runner"] = {"profile": "max"}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def lease(conn, config, profile, now=1.0):
    req = LeaseRequest("node-a", None, False, 0.0, 99.0, "task", profile)
    return lease_job(conn, config, req, now)


def test_an_on_demand_profile_takes_queued_inference_only_when_asked(tmp_path):
    config = make_config(tmp_path, route=True)
    conn = fresh_store(tmp_path / "state")
    submit_job(
        conn,
        config,
        "pa",
        body(privacy="public", input={"messages": [{"role": "user", "content": "hi"}]}),
        0.0,
    )
    assert lease(conn, config, None) is None
    got = lease(conn, config, "max")
    assert got.model == "max"
    assert got.input == {"prompt": "[user]\nhi"}
    assert conn.execute("SELECT model FROM jobs").fetchone()[0] == "model-a"


def test_a_class_the_profile_or_policy_does_not_allow_stays_queued(tmp_path):
    config = make_config(tmp_path, {**MAX, "privacy": ["public", "mail", "secret"]})
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="secret"), 0.0)
    submit_job(conn, config, "pa", body(key="k2", privacy="mail"), 0.0)
    assert lease(conn, config, "max") is None


def test_a_wall_rests_the_profile_so_the_drain_ends(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public"), 0.0)
    submit_job(conn, config, "pa", body(key="k2", privacy="public"), 0.0)
    got = lease(conn, config, "max")
    executor = {"node": "node-a", "provider": "max-lane", "model": "claude-x"}
    report = CompletionReport("failed", None, {}, executor, "quota_wall", 1.0)
    complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0, node="node-a")
    assert lease(conn, config, "max", 3.0) is None


def test_an_unknown_or_automatic_profile_leases_nothing(tmp_path):
    automatic = {"runner": "agy", "privacy": ["public"]}
    config = make_config(tmp_path, automatic)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public"), 0.0)
    assert lease(conn, config, "missing") is None
    assert lease(conn, config, "max") is None


@pytest.mark.parametrize(
    "profile",
    [
        {"runner": "max-lane", "model": "claude-x"},
        {"runner": "max-lane", "on_demand": True},
        {**MAX, "input_root": "."},
    ],
)
def test_a_max_lane_profile_must_be_on_demand_and_tool_free(tmp_path, profile):
    with pytest.raises(ValueError, match="max-lane"):
        make_config(tmp_path, profile)
