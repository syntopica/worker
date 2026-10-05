import json

import pytest

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job

DAY = 86400.0


def make_config(tmp_path, cap):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.bulk"]["local_after_s"] = 600
    raw["queues"]["pa.bulk"]["openrouter"] = {
        "models": {"model-a": "vendor/model:free"},
        "daily_cap": cap,
    }
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def remote(conn, config, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 0.0, "openrouter"), now)


def finish_remotely(conn, config, now):
    lease = remote(conn, config, now)
    report = CompletionReport(
        "succeeded",
        {"text": "{}", "json": {}},
        {"tokens_in": 1, "tokens_out": 1, "cost_usd": 0.0},
        {"node": "node-a", "provider": "openrouter", "model": "vendor/model:free"},
        None,
        1.0,
    )
    complete_attempt(conn, config, lease.attempt_id, lease.generation, report, now + 1)


def test_a_queue_at_its_daily_cap_is_not_leased_to_openrouter(tmp_path):
    config = make_config(tmp_path, 1)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public", key="one"), DAY)
    finish_remotely(conn, config, DAY + 10)
    submit_job(conn, config, "pa", body(privacy="public", key="two"), DAY + 20)

    assert remote(conn, config, DAY + 30) is None


def test_the_cap_resets_at_midnight_utc(tmp_path):
    config = make_config(tmp_path, 1)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public", key="one"), DAY)
    finish_remotely(conn, config, DAY + 10)
    submit_job(conn, config, "pa", body(privacy="public", key="two"), DAY + 20)

    assert remote(conn, config, 2 * DAY + 5) is not None


def test_a_cap_must_be_a_positive_integer(tmp_path):
    with pytest.raises(ValueError, match="daily_cap"):
        make_config(tmp_path, 0)


def test_a_capped_queue_is_not_held_back_from_the_local_model(tmp_path):
    config = make_config(tmp_path, 1)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="public", key="one"), DAY)
    finish_remotely(conn, config, DAY + 10)
    submit_job(conn, config, "pa", body(privacy="public", key="two"), DAY + 20)

    local = lease_job(conn, config, LeaseRequest("node-a", None, False, 64.0, 0.0), DAY + 30)

    assert local is not None
