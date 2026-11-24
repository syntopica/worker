import json
from contextlib import closing

import pytest

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.api_error import ApiError
from worker.jobs.get_job import get_job
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.retry_job import retry_job
from worker.jobs.submit_job import submit_job

ROUTE = {"models": {"model-a": "vendor/model:free"}, "after_s": 0}


def make_config(tmp_path, *, runner=True, openrouter=True, local_after_s=0):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"bulk.agy": {"runner": "agy", "privacy": ["mail"]}}
    queue = raw["queues"]["pa.bulk"]
    queue["local_after_s"] = local_after_s
    if openrouter:
        queue["openrouter"] = ROUTE
    if runner:
        queue["runner"] = {"profile": "bulk.agy"}
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner", "openrouter"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def excluding(*providers, key="k1"):
    requirements = {
        "capability": "chat.json",
        "models": ["model-a"],
        "exclude_providers": providers,
    }
    return body(key=key, requirements=requirements)


def lease(conn, config, kind, now=1.0):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, kind), now)


def state_of(conn, job_id):
    return get_job(conn, "pa", job_id)


def test_an_excluded_runner_never_takes_the_job_but_the_other_rungs_do(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", excluding("agy"), 0.0)
        assert lease(conn, config, "task") is None
        assert lease(conn, config, "openrouter").model == "vendor/model:free"


def test_an_excluded_openrouter_never_takes_the_job(tmp_path):
    config = make_config(tmp_path, runner=False)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", excluding("openrouter"), 0.0)
        assert lease(conn, config, "openrouter") is None
        assert lease(conn, config, "inference", 1e6).model == "model-a"


def test_an_excluded_local_model_never_takes_the_job(tmp_path):
    config = make_config(tmp_path, openrouter=False)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", excluding("ollama"), 0.0)
        assert lease(conn, config, "inference", 1e6) is None
        assert lease(conn, config, "task").model == "bulk.agy"


def test_an_excluded_runner_does_not_hold_the_local_model_back(tmp_path):
    config = make_config(tmp_path, openrouter=False, local_after_s=600)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", body(key="plain"), 0.0)
        assert lease(conn, config, "inference", 10.0) is None
        job_id, _ = submit_job(conn, config, "pa", excluding("agy", key="skip"), 0.0)
        assert lease(conn, config, "inference", 10.0).job_id == job_id


def test_no_field_or_an_empty_list_leaves_the_ladder_unchanged(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", body(key="a"), 0.0)
        submit_job(conn, config, "pa", excluding(key="b"), 0.0)
        assert lease(conn, config, "task").model == "bulk.agy"
        assert lease(conn, config, "task").model == "bulk.agy"
        assert (
            conn.execute(
                "SELECT count(*) FROM jobs WHERE exclude_providers IS NOT NULL"
            ).fetchone()[0]
            == 0
        )


def test_an_exclusion_that_leaves_no_executor_fails_the_job(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, created = submit_job(
            conn, config, "pa", excluding("ollama", "openrouter", "agy"), 0.0
        )
        assert created
        job = state_of(conn, job_id)
        assert (job["state"], job["error"]) == ("failed", "no_eligible_executor")
        assert job["result"]["control"] == "failed"
        assert job["result"]["detail"] == {"error": "no_eligible_executor"}
        assert lease(conn, config, "task") is None
        assert lease(conn, config, "inference", 1e6) is None


def test_the_same_submission_again_returns_the_failed_job(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        raw = excluding("ollama", "openrouter", "agy")
        first, _ = submit_job(conn, config, "pa", raw, 0.0)
        assert submit_job(conn, config, "pa", raw, 1.0) == (first, False)


def test_excluding_a_provider_the_ladder_never_uses_changes_nothing(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", excluding("codex"), 0.0)
        assert state_of(conn, job_id)["state"] == "queued"
        assert lease(conn, config, "task").model == "bulk.agy"


def test_an_exclusion_stored_with_the_job_survives_an_admin_retry(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", excluding("agy"), 0.0)
        conn.execute("UPDATE jobs SET state='cancelled', finished=1.0 WHERE id=?", (job_id,))
        new_id, _ = retry_job(conn, config, job_id, "admin", 2.0)
        stored = conn.execute("SELECT exclude_providers FROM jobs WHERE id=?", (new_id,)).fetchone()
        assert json.loads(stored[0]) == ["agy"]


@pytest.mark.parametrize("value", ["agy", [1], [""], [["agy"]], {"agy": True}, [None]])
def test_a_malformed_exclusion_is_refused(conn, config, value):
    requirements = {"models": ["model-a"], "exclude_providers": value}
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(requirements=requirements), 0.0)
    assert (error.value.status, error.value.code) == (400, "bad_requirements")


def test_an_old_body_without_the_field_is_admitted_unchanged(conn, config):
    job_id, created = submit_job(conn, config, "pa", body(), 0.0)
    assert created
    assert (
        conn.execute("SELECT exclude_providers FROM jobs WHERE id=?", (job_id,)).fetchone()[0]
        is None
    )
