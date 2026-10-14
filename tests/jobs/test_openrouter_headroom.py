import json

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.openrouter_headroom import openrouter_headroom
from worker.jobs.submit_job import submit_job

DAY = 86400.0


def make_config(tmp_path, key_cap):
    raw = json.loads(json.dumps(CONFIG))
    route = {"models": {"model-a": "vendor/model:free"}}
    if key_cap is not None:
        route["daily_key_cap"] = key_cap
    raw["queues"]["pa.bulk"]["openrouter"] = route
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def record(conn, job_id, attempt_id, started):
    conn.execute(
        "INSERT INTO attempts (id, job_id, generation, node, started, provider)"
        " VALUES (?, ?, 1, 'node-a', ?, 'openrouter')",
        (attempt_id, job_id, started),
    )


def test_status_counts_todays_openrouter_attempts_against_the_key_cap(tmp_path):
    config = make_config(tmp_path, 2)
    conn = fresh_store(tmp_path / "state")
    job_id = submit_job(conn, config, "pa", body(), 0.0)[0]
    record(conn, job_id, "old", DAY - 10.0)
    record(conn, job_id, "a1", DAY + 5.0)
    record(conn, job_id, "a2", DAY + 6.0)
    got = openrouter_headroom(config, conn, DAY + 60.0)
    assert got == {"spent_today": 2, "key_cap": 2, "capped": ["pa.bulk"]}


def test_without_a_key_cap_nothing_is_capped(tmp_path):
    config = make_config(tmp_path, None)
    conn = fresh_store(tmp_path / "state")
    got = openrouter_headroom(config, conn, DAY)
    assert got == {"spent_today": 0, "key_cap": None, "capped": []}
