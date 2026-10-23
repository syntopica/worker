import json
from contextlib import closing

from tests.conftest import fresh_store
from tests.ladder.test_on_demand_tasks import setup_task
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def test_a_snapshot_drain_finishes_without_taking_new_producer_work(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        old, _ = submit_job(conn, config, "pa", payload, 0.0)
        new = json.loads(payload)
        new["idempotency_key"] = "new"
        submit_job(conn, config, "pa", json.dumps(new).encode(), 2.0)
        req = LeaseRequest("node-a", None, False, 0.0, 0.0, "task", "max", 1.0)
        got = lease_job(conn, config, req, 3.0)
        assert got.job_id == old
        report = CompletionReport(
            "succeeded",
            {"text": "", "json": {"ok": True}},
            {},
            {"provider": "max-lane", "model": "claude-x"},
            None,
            1.0,
        )
        complete_attempt(conn, config, got.attempt_id, got.generation, report, 4.0)
        assert lease_job(conn, config, req, 5.0) is None
        unbounded = LeaseRequest("node-a", None, False, 0.0, 0.0, "task", "max")
        assert lease_job(conn, config, unbounded, 5.0) is not None
