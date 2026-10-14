import json

from tests.conftest import CONFIG, body, fresh_store
from tests.remote.test_remote_step import LEASE, Link
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job
from worker.remote.run_remote_attempt import run_remote_attempt

ROUTE = {"models": {"model-a": "vendor/model:free"}, "fallbacks": ["vendor/other:free"]}


def make_config(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.bulk"]["openrouter"] = ROUTE
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def sent_body(config, lease):
    bodies = []

    def post(key, sent, timeout):
        bodies.append(sent)
        return None, "transport_error"

    run_remote_attempt(lease, Link(), "node-a", "k", config, lambda: 0.0, post)
    return bodies[0]


def test_a_pinned_remote_job_asks_its_own_model_only(tmp_path):
    config = make_config(tmp_path)
    assert sent_body(config, LEASE)["models"] == ["vendor/model:free", "vendor/other:free"]
    assert "models" not in sent_body(config, {**LEASE, "pinned": True})


def test_the_lease_says_whether_the_job_is_pinned(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.bulk"]["openrouter"] = ROUTE
    raw["queues"]["pa.bulk"]["shadow"] = {
        "rate": 1.0,
        "targets": ["openrouter:vendor/model:free"],
        "judge": "judge.codex",
        "max_pending": 20,
    }
    raw["profiles"] = {"judge.codex": {"runner": "codex", "privacy": ["mail"]}}
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner", "openrouter"], "trust": ["owner"]}}
    (tmp_path / "config.json").write_text(json.dumps(raw))
    config = load_worker_config(tmp_path / "config.json")
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    local = lease_job(
        conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, "inference"), 1.0
    )
    assert local.to_json()["pinned"] is False
    executor = {"node": "node-a", "provider": "ollama", "model": "model-a"}
    done = CompletionReport("succeeded", {"text": "L", "json": None}, {}, executor, None, 1.0)
    complete_attempt(conn, config, local.attempt_id, local.generation, done, 2.0, node="node-a")
    req = LeaseRequest("node-a", None, False, 0.0, 99.0, "openrouter")
    assert lease_job(conn, config, req, 3.0).to_json()["pinned"] is True
