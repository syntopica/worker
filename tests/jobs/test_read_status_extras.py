from worker.jobs.read_status import read_status


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
