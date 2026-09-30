from tests.conftest import body, fresh_store
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job
from worker.jobs.sweep_retention import sweep_retention

EXECUTOR = {"node": "node-a", "provider": "ollama", "model": "model-a"}


def preempt_until_split(conn, config):
    for now in (1.0, 2.0, 3.0):
        leased = lease_job(conn, config, LeaseRequest("node-a", None, False, 40.0, 9999.0), now)
        report = CompletionReport("preempted", None, {}, EXECUTOR, "user_active", 1.0)
        state = complete_attempt(conn, config, leased.attempt_id, leased.generation, report, now)
    return state


def test_an_unanswered_split_request_is_declined_after_the_unacked_ttl(config, tmp_path):
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    assert preempt_until_split(conn, config) == "split_requested"
    ttl = config.queues["pa.bulk"].unacked_ttl_hours * 3600
    sweep_retention(conn, config, 3.0 + ttl - 1)
    assert conn.execute("SELECT state FROM jobs WHERE id=?", (job_id,)).fetchone()[0] == (
        "split_requested"
    )
    sweep_retention(conn, config, 3.0 + ttl + 1)
    row = conn.execute("SELECT state, parked FROM jobs WHERE id=?", (job_id,)).fetchone()
    assert tuple(row) == ("queued", 1)
    control = conn.execute(
        "SELECT acked FROM results WHERE job_id=? AND control='split_requested'", (job_id,)
    ).fetchone()
    assert control[0] is not None
