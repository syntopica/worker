import dataclasses

from tests.conftest import body
from worker.jobs.ack_result import ack_result
from worker.jobs.submit_job import submit_job


def test_declining_a_split_on_a_queue_removed_from_config_parks_the_job(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    conn.execute("UPDATE jobs SET state='split_requested' WHERE id=?", (job_id,))
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, control, created)"
        " VALUES ('r1', ?, 'pa', 'pa.bulk', 'split_requested', 1.0)",
        (job_id,),
    )
    queues = {k: v for k, v in config.queues.items() if k != "pa.bulk"}
    ack_result(conn, dataclasses.replace(config, queues=queues), "pa", job_id, "r1", True, 2.0)
    row = conn.execute("SELECT state, parked FROM jobs WHERE id=?", (job_id,)).fetchone()
    assert (row["state"], row["parked"]) == ("queued", 1)
