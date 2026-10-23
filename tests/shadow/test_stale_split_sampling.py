from tests.conftest import body
from worker.jobs.cancel_stale_shadows import cancel_stale_shadows
from worker.jobs.get_job import get_job
from worker.jobs.submit_job import submit_job


def test_stale_sampling_split_requests_close_without_cancelling_production(conn, config):
    shadow, _ = submit_job(conn, config, "pa", body(key="shadow"), 0.0)
    production, _ = submit_job(conn, config, "pa", body(key="production"), 0.0)
    conn.execute("UPDATE jobs SET state='split_requested'")
    conn.execute("UPDATE jobs SET producer='_shadow' WHERE id=?", (shadow,))
    assert cancel_stale_shadows(conn, 86401.0) == 1
    assert conn.execute("SELECT state FROM jobs WHERE id=?", (shadow,)).fetchone()[0] == "cancelled"
    assert get_job(conn, "pa", production)["state"] == "split_requested"
