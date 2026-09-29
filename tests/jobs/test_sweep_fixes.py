import json
import time

from tests.conftest import body
from tests.jobs.test_results_and_retention import succeed
from worker.jobs.ack_result import ack_result
from worker.jobs.check_outstanding import check_outstanding
from worker.jobs.list_results import list_results
from worker.jobs.submit_job import submit_job
from worker.jobs.sweep_retention import sweep_retention
from worker.store.open_store import open_store

DAY = 86400.0
_WRITES = ("UPDATE", "DELETE", "INSERT")


def ack_all(conn, config, now):
    for r in list_results(conn, "pa", "pa.live", 0, 100):
        ack_result(conn, config, "pa", r["job_id"], r["result_id"], False, now)


def test_a_sweep_over_already_swept_jobs_touches_no_rows(conn, config):
    for i in range(20):
        submit_job(conn, config, "pa", body(key=f"k{i}", queue="pa.live"), 0.0)
    conn.execute("UPDATE jobs SET state='failed', finished=1, updated=1")
    sweep_retention(conn, config, 2 * DAY)  # sensitive: payloads deleted after inspection
    writes = []
    conn.set_trace_callback(
        lambda sql: writes.append(sql) if sql.lstrip().upper().startswith(_WRITES) else None
    )
    before = conn.total_changes
    sweep_retention(conn, config, 2 * DAY + 60)
    conn.set_trace_callback(None)
    assert (conn.total_changes - before, writes) == (0, [])


def seed(state, count):
    """``count`` public jobs, succeeded and acked at t=0, each with input and output."""
    conn = open_store(state)
    conn.execute("BEGIN")
    for i in range(count):
        job, result = f"j{i}", f"r{i}"
        conn.execute(
            "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority,"
            " privacy, model, state, max_attempts, not_before, created, updated, finished, acked)"
            " VALUES (?, 'pa', 'pa.live', 'inference', ?, 'h', 50, 'public', 'model-a', 'succeeded',"
            " 3, 0, 0, 0, 0, 0)",
            (job, job),
        )
        conn.execute(
            "INSERT INTO results (result_id, job_id, producer, queue, created, acked)"
            " VALUES (?, ?, 'pa', 'pa.live', 0, 0)",
            (result, job),
        )
        conn.execute("INSERT INTO p.inputs VALUES (?, ?)", (job, json.dumps({"input": {}})))
        conn.execute("INSERT INTO p.outputs VALUES (?, ?)", (result, "{}"))
    conn.execute("COMMIT")
    return conn


def timed_sweep(state, config, count):
    conn = seed(state, count)
    started = time.perf_counter()
    swept = sweep_retention(conn, config, 8 * DAY)
    elapsed = time.perf_counter() - started
    assert swept == count
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 0
    conn.close()
    return elapsed


def test_the_sweep_is_linear_and_fast_at_five_thousand_jobs(tmp_path, config):
    small = timed_sweep(tmp_path / "small", config, 1250)
    large = timed_sweep(tmp_path / "large", config, 5000)
    assert large < 5.0
    assert large / max(small, 0.05) < 8  # quadratic would be about 16


def test_unacked_public_output_expires_with_a_control_result(conn, config):
    job_id = succeed(conn, config, privacy="public", queue="pa.live")
    sweep_retention(conn, config, 6 * DAY)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "succeeded"
    sweep_retention(conn, config, 8 * DAY)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "unacked_expired"
    assert conn.execute("SELECT count(*) FROM p.outputs").fetchone()[0] == 0
    feed = list_results(conn, "pa", "pa.live", 0, 100)
    assert [(r["job_id"], r["control"]) for r in feed] == [(job_id, "unacked_expired")]
    check_outstanding(conn, config, "pa", "pa.live")  # the expired job holds no slot
    assert (
        conn.execute(
            "SELECT count(*) FROM jobs WHERE producer='pa' AND queue='pa.live' AND acked IS NULL"
            " AND state NOT IN ('superseded','cancelled','unacked_expired')"
        ).fetchone()[0]
        == 0
    )


def test_the_feed_never_shows_a_result_whose_output_is_gone(conn, config):
    succeed(conn, config, privacy="public", queue="pa.live")
    conn.execute("DELETE FROM p.outputs")
    assert list_results(conn, "pa", "pa.live", 0, 100) == []


def test_limit_is_clamped_to_one_through_a_hundred(conn, config):
    for i in range(3):
        succeed(conn, config, key=f"k{i}", privacy="public", queue="pa.live")
    assert len(list_results(conn, "pa", "pa.live", 0, -1)) == 1
    assert len(list_results(conn, "pa", "pa.live", 0, 0)) == 1


def test_a_released_key_creates_a_new_job(conn, config):
    job_id = succeed(conn, config, key="same", privacy="public", queue="pa.live")
    ack_all(conn, config, 10.0)
    again, created = submit_job(
        conn, config, "pa", body(key="same", queue="pa.live", privacy="public"), 20.0
    )
    assert (again, created) == (job_id, False)  # inside the window: the same job
    sweep_retention(conn, config, 10.0 + 7 * DAY + 1)
    for table in ("jobs", "results", "attempts"):
        assert conn.execute(f"SELECT count(*) FROM {table}").fetchone()[0] == 0, table  # noqa: S608
    fresh, created = submit_job(
        conn, config, "pa", body(key="same", queue="pa.live", privacy="public"), 10.0 + 7 * DAY + 2
    )
    assert created
    assert fresh != job_id


def test_an_unacked_failure_is_released_after_retention(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(queue="pa.live"), 0.0)
    conn.execute("UPDATE jobs SET state='failed', finished=1, updated=1 WHERE id=?", (job_id,))
    sweep_retention(conn, config, 2 * DAY)  # sensitive: payload gone after inspection
    assert conn.execute("SELECT payloads_deleted FROM jobs").fetchone()[0] == 1
    sweep_retention(conn, config, 1 + 7 * DAY + 1)
    assert conn.execute("SELECT count(*) FROM jobs").fetchone()[0] == 0


def test_an_unknown_queue_gets_default_retention_and_one_log_line(conn, config, capsys):
    succeed(conn, config, privacy="public", queue="pa.live")
    conn.execute("UPDATE jobs SET queue='gone.queue'")
    conn.execute("UPDATE results SET queue='gone.queue'")
    sweep_retention(conn, config, 8 * DAY)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "unacked_expired"
    err = capsys.readouterr().err
    assert err.count("\n") == 1
    assert "gone.queue" in err


def test_the_payload_wal_is_checkpointed_after_a_deleting_pass(tmp_path, config):
    state = tmp_path / "state"
    conn = open_store(state)
    succeed(conn, config, privacy="public", queue="pa.live")
    sweep_retention(conn, config, 8 * DAY)
    assert (state / "payloads.sqlite3-wal").stat().st_size == 0
