from tests.conftest import body
from worker.jobs.read_activity import read_activity
from worker.jobs.submit_job import submit_job

HOUR = 3600.0
NOW = 100 * HOUR + 1800.0


def attempt(conn, job_id, n, ended, *, outcome="succeeded", error=None, provider="ollama"):
    conn.execute(
        "INSERT INTO attempts (id, job_id, generation, node, started, ended, outcome, error,"
        " wall_s, tokens_in, tokens_out, provider)"
        " VALUES (?, ?, 1, 'node-a', ?, ?, ?, ?, 10.0, 5, 7, ?)",
        (f"a{n}", job_id, ended - 10.0, ended, outcome, error, provider),
    )


def test_activity_groups_attempts_by_hour_queue_provider_and_outcome(conn, config):
    job, _ = submit_job(conn, config, "pa", body(key="k1"), 0.0)
    attempt(conn, job, 1, 99 * HOUR + 60.0)
    attempt(conn, job, 2, 99 * HOUR + 120.0)
    attempt(
        conn, job, 3, 99 * HOUR + 180.0, outcome="failed", error="no_output", provider="openrouter"
    )
    attempt(conn, job, 4, 50 * HOUR, provider=None)
    activity = read_activity(conn, NOW, 24)
    assert activity["bucket_s"] == HOUR
    assert activity["since"] == NOW - 24 * HOUR
    assert activity["rows"] == [
        {
            "bucket": 99 * HOUR,
            "queue": "pa.bulk",
            "provider": "ollama",
            "sampling": False,
            "outcome": "succeeded",
            "error": None,
            "attempts": 2,
            "wall_s": 20.0,
            "tokens_in": 10,
            "tokens_out": 14,
        },
        {
            "bucket": 99 * HOUR,
            "queue": "pa.bulk",
            "provider": "openrouter",
            "sampling": False,
            "outcome": "failed",
            "error": "no_output",
            "attempts": 1,
            "wall_s": 10.0,
            "tokens_in": 5,
            "tokens_out": 7,
        },
    ]


def test_activity_marks_shadow_and_judge_attempts_as_sampling(conn, config):
    job, _ = submit_job(conn, config, "pa", body(key="k1"), 0.0)
    conn.execute("UPDATE jobs SET producer='_judge' WHERE id=?", (job,))
    attempt(conn, job, 1, 99 * HOUR, outcome="failed", error="runner_failed", provider=None)
    (row,) = read_activity(conn, NOW, 24)["rows"]
    assert row["sampling"] is True
    assert row["provider"] == "unknown"


def test_a_week_uses_six_hour_buckets_and_hours_are_clamped(conn):
    assert read_activity(conn, NOW, 168)["bucket_s"] == 6 * HOUR
    assert read_activity(conn, NOW, 10_000)["since"] == NOW - 168 * HOUR
    assert read_activity(conn, NOW, 0)["since"] == NOW - HOUR
