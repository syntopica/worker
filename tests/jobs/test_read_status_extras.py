from tests.conftest import body
from worker.jobs.read_status import read_status
from worker.jobs.submit_job import submit_job


def test_status_shows_runner_cooldowns_and_each_node_last_release(conn):
    conn.execute("INSERT INTO nodes (name, report, updated) VALUES ('node-a', '{}', 90.0)")
    conn.execute(
        "INSERT INTO attempts (id, job_id, generation, node, started, ended, outcome, error)"
        " VALUES ('a1', 'j', 1, 'node-a', 10.0, 20.0, 'preempted', 'user_active'),"
        " ('a2', 'j', 2, 'node-a', 30.0, 40.0, 'preempted', 'memory_pressure')"
    )
    conn.execute("INSERT INTO cooldowns (runner, until) VALUES ('codex', 700.0), ('agy', 50.0)")
    status = read_status(conn, 100.0)
    assert status["nodes"]["node-a"]["last_release"] == {"code": "memory_pressure", "age_s": 60.0}
    assert status["cooldowns"] == {"codex": 600.0}


def test_recent_failures_leave_out_shadow_and_judge_jobs(conn, config):
    ids = [submit_job(conn, config, "pa", body(key=f"k{i}"), float(i))[0] for i in range(3)]
    for job_id, producer in zip(ids, ["pa", "_shadow", "_judge"], strict=True):
        conn.execute(
            "UPDATE jobs SET state='failed', producer=?, error='no_output', finished=10.0"
            " WHERE id=?",
            (producer, job_id),
        )
    failures = read_status(conn, 20.0)["recent_failures"]
    assert [f["id"] for f in failures] == [ids[0]]


def test_each_queue_counts_its_failed_sampling_jobs(conn, config):
    ids = [submit_job(conn, config, "pa", body(key=f"k{i}"), float(i))[0] for i in range(3)]
    for job_id, producer in zip(ids, ["pa", "_shadow", "_judge"], strict=True):
        conn.execute("UPDATE jobs SET state='failed', producer=? WHERE id=?", (producer, job_id))
    queue = read_status(conn, 20.0)["queues"]["pa.bulk"]
    assert queue["states"]["failed"] == 3
    assert queue["sampling_failed"] == 2


def test_queue_age_counts_production_jobs_and_sampling_is_reported_apart(conn, config):
    ids = [submit_job(conn, config, "pa", body(key=f"k{i}"), float(i))[0] for i in range(3)]
    for job_id, producer in zip(ids, ["_shadow", "pa", "_judge"], strict=True):
        conn.execute("UPDATE jobs SET producer=? WHERE id=?", (producer, job_id))
    queue = read_status(conn, 20.0)["queues"]["pa.bulk"]
    assert queue["states"]["queued"] == 3
    assert queue["sampling_queued"] == 2
    assert queue["oldest_queued_s"] == 19.0


def test_a_queue_with_only_sampling_jobs_queued_has_no_age(conn, config):
    job_id = submit_job(conn, config, "pa", body(key="k"), 1.0)[0]
    conn.execute("UPDATE jobs SET producer='_shadow' WHERE id=?", (job_id,))
    queue = read_status(conn, 20.0)["queues"]["pa.bulk"]
    assert queue["sampling_queued"] == 1
    assert queue["oldest_queued_s"] is None
