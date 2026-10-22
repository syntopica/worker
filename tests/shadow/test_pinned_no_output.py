import json
from contextlib import closing

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def make_config(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {
        "bulk.agy": {"runner": "agy", "privacy": ["mail"]},
        "judge.codex": {"runner": "codex", "privacy": ["mail"]},
    }
    raw["queues"]["pa.bulk"]["shadow"] = {
        "rate": 1.0,
        "targets": ["runner:bulk.agy"],
        "judge": "judge.codex",
        "max_pending": 20,
    }
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner", "openrouter"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def lease(conn, config, kind, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, kind), now)


def report(provider, outcome, output, error):
    executor = {"node": "node-a", "provider": provider, "model": ""}
    return CompletionReport(outcome, output, {}, executor, error, 1.0)


def test_an_empty_answer_to_a_pinned_member_is_charged_not_retried_forever(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", body(), 0.0)
        got = lease(conn, config, "inference", 1.0)
        answer = {"text": "L", "json": None}
        done = report("ollama", "succeeded", answer, None)
        complete_attempt(conn, config, got.attempt_id, got.generation, done, 2.0, node="node-a")
        member = lease(conn, config, "task", 3.0)
        assert member.model == "bulk.agy"
        empty = report("agy", "failed", None, "no_output")
        complete_attempt(
            conn, config, member.attempt_id, member.generation, empty, 4.0, node="node-a"
        )
        row = conn.execute(
            "SELECT state, error, attempts FROM jobs WHERE pin=?", ("runner:bulk.agy",)
        )
        assert row.fetchone()[:] == ("failed", "no_output", 1)
        assert lease(conn, config, "task", 5.0) is None
