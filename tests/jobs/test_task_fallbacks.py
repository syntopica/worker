import json

import pytest

from tests.conftest import CONFIG, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.job_cooling_until import job_cooling_until
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def make_config(tmp_path, fallbacks, agy_privacy=("internal",)):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.syn"] = {
        "run_when": "idle",
        "profiles": ["pa.cursor", "pa.agy"],
        "fallbacks": fallbacks,
    }
    raw["producers"]["pa"].append("pa.syn")
    raw["profiles"] = {
        "pa.cursor": {"runner": "cursor", "privacy": ["internal"]},
        "pa.agy": {"runner": "agy", "privacy": list(agy_privacy)},
    }
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def task(key="t1"):
    body = {
        "contract": 1,
        "kind": "task",
        "queue": "pa.syn",
        "idempotency_key": key,
        "privacy": "internal",
        "input": {"profile": "pa.cursor", "prompt": "summarise"},
    }
    return json.dumps(body).encode()


def lease(conn, config, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 9999.0, "task"), now)


def wall(conn, config, leased, now):
    executor = {"node": "node-a", "provider": "runner", "model": ""}
    report = CompletionReport("failed", None, {}, executor, "quota_wall", 1.0)
    return complete_attempt(conn, config, leased.attempt_id, leased.generation, report, now)


def test_a_walled_task_runs_at_once_under_the_queue_fallback(tmp_path):
    config = make_config(tmp_path, {"pa.cursor": ["pa.agy"]})
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", task(), 0.0)
    first = lease(conn, config, 1.0)
    assert first.model == "pa.cursor"
    assert wall(conn, config, first, 2.0) == "queued"
    assert job_cooling_until(conn, config, job_id, 3.0) is None
    second = lease(conn, config, 3.0)
    assert (second.job_id, second.model) == (job_id, "pa.agy")


def test_without_a_fallback_the_task_waits_out_the_cooldown(tmp_path):
    config = make_config(tmp_path, {})
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", task(), 0.0)
    wall(conn, config, lease(conn, config, 1.0), 2.0)
    assert lease(conn, config, 3.0) is None
    assert job_cooling_until(conn, config, job_id, 3.0) == 3602.0


def test_a_fallback_narrower_than_its_primary_is_refused(tmp_path):
    with pytest.raises(ValueError, match="privacy"):
        make_config(tmp_path, {"pa.cursor": ["pa.agy"]}, agy_privacy=("public",))


def test_a_fallback_must_be_granted_by_its_queue(tmp_path):
    with pytest.raises(ValueError, match="granted"):
        make_config(tmp_path, {"pa.cursor": ["pa.other"]})
