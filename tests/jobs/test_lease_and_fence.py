import pytest

from tests.conftest import body
from worker.jobs.api_error import ApiError
from worker.jobs.expire_leases import expire_leases
from worker.jobs.heartbeat_attempt import heartbeat_attempt
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def ask(node="node-a", active=False):
    return LeaseRequest(node, None, active, 44.0, 900.0)


def test_lease_mints_attempt_and_generation(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    lease = lease_job(conn, config, ask(), 1.0)
    assert (lease.job_id, lease.generation, lease.input["messages"][0]["content"]) == (
        job_id,
        1,
        "hi",
    )
    assert lease_job(conn, config, ask(), 2.0) is None


def test_guest_node_gets_no_mail(conn, config):
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease_job(conn, config, ask(node="node-g"), 1.0) is None


def test_expired_lease_is_reissued_and_the_old_attempt_is_fenced(conn, config):
    submit_job(conn, config, "pa", body(), 0.0)
    first = lease_job(conn, config, ask(), 1.0)
    assert expire_leases(conn, 1.0 + 61) == 1
    second = lease_job(conn, config, ask(), 70.0)
    assert second.generation == first.generation + 1
    with pytest.raises(ApiError) as error:
        heartbeat_attempt(conn, first.attempt_id, first.generation, False, 71.0)
    assert error.value.code == "stale_attempt"
    heartbeat_attempt(conn, second.attempt_id, second.generation, False, 71.0)


def test_lost_leases_count_as_attempts_and_end_in_a_failed_control_result(conn, config):
    submit_job(conn, config, "pa", body(max_attempts=1), 0.0)
    lease_job(conn, config, ask(), 1.0)
    expire_leases(conn, 100.0)
    row = conn.execute("SELECT state, error FROM jobs").fetchone()
    assert (row["state"], row["error"]) == ("failed", "lease_lost")
    assert conn.execute("SELECT control FROM results").fetchone()[0] == "failed"


def test_deadline_expires_a_queued_job(conn, config):
    submit_job(conn, config, "pa", body(deadline=50.0), 0.0)
    expire_leases(conn, 60.0)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "expired"
    assert conn.execute("SELECT control FROM results").fetchone()[0] == "expired"
