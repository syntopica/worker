import json

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def make_config(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {
        "a.pro": {"runner": "agy", "model": "pro", "privacy": ["mail"]},
        "b.flash": {"runner": "agy", "model": "flash", "privacy": ["mail"]},
    }
    raw["queues"]["pa.bulk"]["runner"] = {"profile": "a.pro"}
    raw["queues"]["pa.live"]["runner"] = {"profile": "b.flash"}
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def lease(conn, config, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, "task"), now)


def test_a_wall_on_one_model_rests_that_model_only(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    got = lease(conn, config, 1.0)
    assert got.model == "a.pro"
    executor = {"node": "node-a", "provider": "agy", "model": "pro"}
    report = CompletionReport("failed", None, {}, executor, "quota_wall", 1.0)
    complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0, node="node-a")
    assert [r[0] for r in conn.execute("SELECT runner FROM cooldowns")] == ["agy:pro"]
    submit_job(conn, config, "pa", body(key="k2", queue="pa.live"), 3.0)
    assert lease(conn, config, 4.0).model == "b.flash"
    assert lease(conn, config, 5.0) is None
