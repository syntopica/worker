import json

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job

ROUTE = {"models": {"model-a": "vendor/model:free"}, "after_s": 100}


def make_config(tmp_path, *, runner=True, local_after_s=600):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"bulk.agy": {"runner": "agy", "privacy": ["public", "mail"]}}
    queue = raw["queues"]["pa.bulk"]
    queue["openrouter"] = ROUTE
    queue["local_after_s"] = local_after_s
    if runner:
        queue["runner"] = {"profile": "bulk.agy"}
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner", "openrouter"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def lease(conn, config, kind, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, kind), now)


def test_the_runner_takes_inference_first_as_a_rendered_task(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    schema = {"type": "object"}
    job_input = {"messages": [{"role": "user", "content": "hi"}], "schema": schema}
    submit_job(conn, config, "pa", body(input=job_input), 0.0)
    assert lease(conn, config, "inference", 1.0) is None
    assert lease(conn, config, "openrouter", 1.0) is None
    got = lease(conn, config, "task", 1.0)
    assert got.model == "bulk.agy"
    assert got.input == {"prompt": "[user]\nhi", "output_schema": schema}
    assert conn.execute("SELECT model FROM jobs").fetchone()[0] == "model-a"


def test_openrouter_waits_for_the_runner_then_local_after_that(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, "openrouter", 50.0) is None
    assert lease(conn, config, "openrouter", 150.0).model == "vendor/model:free"


def test_a_resting_runner_sends_the_job_to_openrouter_at_once(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    conn.execute("INSERT INTO cooldowns (runner, until) VALUES ('agy', 1e12)")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, "task", 1.0) is None
    assert lease(conn, config, "openrouter", 1.0) is not None


def test_local_takes_a_delayed_job_or_one_no_remote_may_see(tmp_path):
    config = make_config(tmp_path, runner=False)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, "inference", 10.0) is None
    assert lease(conn, config, "inference", 700.0) is not None
    submit_job(conn, config, "pa", body(key="k2", privacy="secret"), 0.0)
    assert lease(conn, config, "inference", 1.0) is not None


def test_a_runner_wall_rests_it_and_frees_the_job_for_the_next_rung(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    got = lease(conn, config, "task", 1.0)
    executor = {"node": "node-a", "provider": "agy", "model": ""}
    report = CompletionReport("failed", None, {}, executor, "quota_wall", 1.0)
    complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0, node="node-a")
    assert conn.execute("SELECT state, attempts FROM jobs").fetchone()[:] == ("queued", 0)
    assert lease(conn, config, "openrouter", 3.0) is not None


def test_a_runner_pinned_to_another_node_holds_no_other_rung_back(tmp_path):
    config = make_config(tmp_path)
    raw = json.loads((tmp_path / "config.json").read_text())
    raw["profiles"]["bulk.agy"]["nodes"] = ["node-b"]
    (tmp_path / "config.json").write_text(json.dumps(raw))
    config = load_worker_config(tmp_path / "config.json")
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, "task", 1.0) is None
    assert lease(conn, config, "openrouter", 1.0) is not None
