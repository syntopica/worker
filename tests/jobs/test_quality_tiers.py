import json

import pytest

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.ack_result import ack_result
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.read_attempt_quality import read_attempt_quality
from worker.jobs.read_result_ratings import read_result_ratings
from worker.jobs.submit_job import submit_job

SCHEMA = {"type": "object", "required": ["a"]}


def make_config(tmp_path, tiers, agy_privacy=("internal",), queue="pa.bulk"):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.syn"] = {"profiles": ["pa.cursor", "pa.agy"]}
    raw["queues"][queue]["tiers"] = tiers
    raw["producers"]["pa"].append("pa.syn")
    raw["profiles"] = {
        "pa.cursor": {"runner": "cursor", "privacy": ["internal"]},
        "pa.agy": {"runner": "agy", "privacy": list(agy_privacy)},
    }
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def task(tier):
    return json.dumps(
        {
            "contract": 1,
            "kind": "task",
            "queue": "pa.syn",
            "idempotency_key": f"t-{tier}",
            "privacy": "internal",
            "tier": tier,
            "input": {"profile": "pa.cursor", "prompt": "summarise"},
        }
    ).encode()


def model_of(conn, job_id):
    return conn.execute("SELECT model, tier FROM jobs WHERE id=?", (job_id,)).fetchone()


def test_a_strong_job_prefers_the_tier_models_and_a_basic_one_the_producers(tmp_path):
    config = make_config(tmp_path, {"strong": {"models": ["model-b"]}})
    conn = fresh_store(tmp_path / "state")
    strong, _ = submit_job(conn, config, "pa", body("k1", tier="strong"), 0.0)
    basic, _ = submit_job(conn, config, "pa", body("k2"), 0.0)
    assert tuple(model_of(conn, strong)) == ("model-b", "strong")
    assert tuple(model_of(conn, basic)) == ("model-a", "basic")


def test_a_tier_the_queue_does_not_map_behaves_as_basic(tmp_path):
    config = make_config(tmp_path, {})
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", body("k1", tier="strong"), 0.0)
    assert tuple(model_of(conn, job_id)) == ("model-a", "strong")


def test_an_unknown_tier_is_refused(tmp_path):
    config = make_config(tmp_path, {})
    conn = fresh_store(tmp_path / "state")
    with pytest.raises(ApiError) as caught:
        submit_job(conn, config, "pa", body("k1", tier="pro"), 0.0)
    assert caught.value.code == "unknown_tier"


def test_a_strong_task_runs_under_the_tier_profile(tmp_path):
    config = make_config(
        tmp_path, {"strong": {"profiles": {"pa.cursor": "pa.agy"}}}, queue="pa.syn"
    )
    conn = fresh_store(tmp_path / "state")
    strong, _ = submit_job(conn, config, "pa", task("strong"), 0.0)
    basic, _ = submit_job(conn, config, "pa", task("basic"), 0.0)
    assert model_of(conn, strong)["model"] == "pa.agy"
    assert model_of(conn, basic)["model"] == "pa.cursor"


@pytest.mark.parametrize(
    ("tiers", "match"),
    [
        ({"strong": {"models": ["model-z"]}}, "unknown models"),
        ({"strong": {"profiles": {"pa.cursor": "pa.agy"}}}, "privacy"),
        ({"strong": {"profiles": {"pa.cursor": "pa.none"}}}, "granted"),
        ({"pro": {}}, "must be one of"),
        ({"strong": {"models": "model-b"}}, "models list"),
        ([], "object"),
    ],
)
def test_a_bad_tier_route_is_refused_at_load(tmp_path, tiers, match):
    with pytest.raises(ValueError, match=match):
        make_config(tmp_path, tiers, agy_privacy=("public",), queue="pa.syn")


def settle(conn, config, now, output, model="model-a"):
    leased = lease_job(conn, config, LeaseRequest("node-a", None, False, 40.0, 9999.0), now)
    executor = {"node": "node-a", "provider": "ollama", "model": model}
    report = CompletionReport("succeeded", output, {}, executor, None, 2.0)
    complete_attempt(conn, config, leased.attempt_id, leased.generation, report, now + 1)
    return leased.job_id


def test_quality_counts_outcomes_and_ratings_by_model(tmp_path):
    config = make_config(tmp_path, {})
    conn = fresh_store(tmp_path / "state")
    message = [{"role": "user", "content": "hi"}]
    extra = {"input": {"messages": message, "schema": SCHEMA}, "max_attempts": 1}
    submit_job(conn, config, "pa", body("k1", **extra), 0.0)
    submit_job(conn, config, "pa", body("k2", **extra), 0.0)
    good = settle(conn, config, 10.0, {"json": {"a": 1}})
    settle(conn, config, 20.0, {"json": {"b": 1}})
    [row] = read_attempt_quality(conn, 0.0)
    assert (row["model"], row["attempts"], row["succeeded"]) == ("model-a", 2, 1)
    assert (row["schema_violations"], row["failed"], row["mean_wall_s"]) == (1, 0, 2.0)
    result_id = conn.execute("SELECT result_id FROM results WHERE job_id=?", (good,)).fetchone()[0]
    ack_result(conn, config, "pa", good, result_id, False, 30.0, "edited")
    [rated] = read_result_ratings(conn, 0.0)
    assert (rated["model"], rated["results"], rated["rated"], rated["edited"]) == (
        "model-a",
        1,
        1,
        1,
    )
