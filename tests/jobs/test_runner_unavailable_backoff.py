from tests.conftest import body
from worker.jobs.fail_attempt import fail_attempt
from worker.jobs.submit_job import submit_job


def job_row(conn, job_id):
    return conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()


def test_an_unavailable_runner_is_retried_after_its_outage_not_within_it(conn, config):
    job_id = submit_job(conn, config, "pa", body(), 0.0)[0]
    conn.execute("UPDATE jobs SET max_attempts=2 WHERE id=?", (job_id,))
    assert fail_attempt(conn, job_row(conn, job_id), "runner_unavailable", 100.0) == "queued"
    assert job_row(conn, job_id)["not_before"] == 700.0


def test_other_failures_keep_the_short_backoff(conn, config):
    job_id = submit_job(conn, config, "pa", body(), 0.0)[0]
    conn.execute("UPDATE jobs SET max_attempts=2 WHERE id=?", (job_id,))
    fail_attempt(conn, job_row(conn, job_id), "runner_failed", 100.0)
    assert job_row(conn, job_id)["not_before"] == 130.0
