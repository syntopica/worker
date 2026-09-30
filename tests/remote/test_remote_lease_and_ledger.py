import json

import pytest

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.read_costs import read_costs
from worker.jobs.submit_job import submit_job

ROUTE = {"models": {"model-a": "vendor/model:free"}, "after_s": 100}


def make_config(tmp_path, route=ROUTE):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.bulk"]["openrouter"] = route
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def remote(conn, config, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 0.0, "openrouter"), now)


def test_a_public_job_escalates_only_after_waiting(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", body(privacy="public"), 0.0)
    assert remote(conn, config, 50.0) is None
    lease = remote(conn, config, 150.0)
    assert (lease.job_id, lease.model, lease.queue, lease.privacy) == (
        job_id,
        "vendor/model:free",
        "pa.bulk",
        "public",
    )
    assert conn.execute("SELECT model FROM jobs").fetchone()[0] == "model-a"


def test_a_mail_job_never_escalates_by_default(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="mail"), 0.0)
    assert remote(conn, config, 500.0) is None


def test_only_free_endpoints_are_accepted(tmp_path):
    with pytest.raises(ValueError, match=":free"):
        make_config(tmp_path, {"models": {"model-a": "vendor/paid"}})


def test_the_ledger_records_provider_and_cost_per_day(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public"), 0.0)
    lease = remote(conn, config, 200.0)
    report = CompletionReport(
        "succeeded",
        {"text": "{}", "json": {}},
        {"tokens_in": 7, "tokens_out": 3, "cost_usd": 0.5},
        {"node": "node-a", "provider": "openrouter", "model": "vendor/model:free"},
        None,
        2.0,
    )
    complete_attempt(conn, config, lease.attempt_id, lease.generation, report, 210.0)
    rows = read_costs(conn, 0.0)
    assert rows == [
        {
            "provider": "openrouter",
            "queue": "pa.bulk",
            "day": "1970-01-01",
            "attempts": 1,
            "succeeded": 1,
            "tokens_in": 7,
            "tokens_out": 3,
            "cost_usd": 0.5,
            "wall_s": 2.0,
        }
    ]
