import json
from contextlib import closing

import pytest

from tests.conftest import CONFIG, fresh_store
from tests.ladder.test_on_demand_profile import MAX, lease
from worker.config.load_worker_config import load_worker_config
from worker.jobs.advance_shadow_groups import advance_shadow_groups
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.get_job import get_job
from worker.jobs.insert_shadow_job import insert_shadow_job
from worker.jobs.read_judgements import read_judgements
from worker.jobs.submit_job import submit_job


def setup_task(tmp_path, *, inputs=False, source=None, cap=None):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {"max": MAX, "source": source or {"runner": "codex", "privacy": ["public"]}}
    if inputs:
        raw["profiles"]["source"]["input_root"] = "."
    raw["queues"]["pa.bulk"]["profiles"] = ["source"]
    if cap:
        raw["runners"] = {"max-lane": {"max_concurrent": cap}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    config = load_worker_config(path)
    payload = json.dumps(
        {
            "contract": 1,
            "kind": "task",
            "queue": "pa.bulk",
            "privacy": "public",
            "idempotency_key": "task",
            "input": {
                "profile": "source",
                "prompt": "Return the result.",
                "inputs": ["file.txt"] if inputs else [],
                "output_schema": {
                    "type": "object",
                    "required": ["ok"],
                    "properties": {"ok": {"type": "boolean"}},
                },
            },
        }
    ).encode()
    return config, payload


def test_on_demand_task_keeps_its_original_profile_and_schema(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", payload, 0.0)
        got = lease(conn, config, "max")
        assert got is not None
        assert got.model == "max"
        assert (
            conn.execute("SELECT model FROM jobs WHERE id=?", (job_id,)).fetchone()[0] == "source"
        )
        bad = CompletionReport(
            "succeeded",
            {"text": "{}", "json": {}},
            {},
            {"provider": "max-lane", "model": "claude-x"},
            None,
            1.0,
        )
        assert complete_attempt(conn, config, got.attempt_id, got.generation, bad, 2.0) == "queued"


def test_on_demand_task_wall_rests_only_the_override_and_preserves_fencing(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", payload, 0.0)
        got = lease(conn, config, "max")
        assert got is not None
        report = CompletionReport(
            "failed", None, {}, {"provider": "max-lane", "model": "claude-x"}, "quota_wall", 1.0
        )
        assert (
            complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0) == "queued"
        )
        assert lease(conn, config, "max", 3.0) is None
        assert lease(conn, config, None, 3.0).model == "source"
        with pytest.raises(ApiError, match="stale_attempt"):
            complete_attempt(conn, config, got.attempt_id, got.generation, report, 4.0)


def test_on_demand_refuses_input_files(tmp_path):
    config, payload = setup_task(tmp_path, inputs=True)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", payload, 0.0)
        assert lease(conn, config, "max") is None


@pytest.mark.parametrize(
    "source", [{"runner": "codex", "privacy": ["public"], "nodes": ["node-g"]}]
)
def test_on_demand_preserves_source_node_restrictions(tmp_path, source):
    config, payload = setup_task(tmp_path, source=source)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", payload, 0.0)
        assert lease(conn, config, "max") is None


def test_on_demand_concurrency_counts_the_override(tmp_path):
    config, payload = setup_task(tmp_path, cap=1)
    with closing(fresh_store(tmp_path / "state")) as conn:
        submit_job(conn, config, "pa", payload, 0.0)
        other = json.loads(payload)
        other["idempotency_key"] = "other"
        submit_job(conn, config, "pa", json.dumps(other).encode(), 0.0)
        assert lease(conn, config, "max") is not None
        assert lease(conn, config, "max", 2.0) is None


@pytest.mark.parametrize(
    "pin,allowed", [("judge", True), ("runner:source", False), ("openrouter:model-a", False)]
)
def test_on_demand_can_judge_but_cannot_replace_comparison_targets(tmp_path, pin, allowed):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", payload, 0.0)
        source = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
        conn.execute("UPDATE jobs SET state='cancelled' WHERE id=?", (job_id,))
        insert_shadow_job(
            conn,
            source,
            {"prompt": "Judge.", "labels": {"A": {"provider": "other", "model": "other"}}},
            group=job_id,
            pin=pin,
            model="source",
            kind="task" if pin == "judge" else "inference",
            producer="_judge" if pin == "judge" else "_shadow",
            now=0.0,
        )
        assert (lease(conn, config, "max") is not None) is allowed


def test_on_demand_judge_cannot_grade_its_own_executor(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", payload, 0.0)
        source = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
        conn.execute("UPDATE jobs SET state='cancelled' WHERE id=?", (job_id,))
        insert_shadow_job(
            conn,
            source,
            {"prompt": "Judge.", "labels": {"A": {"provider": "max-lane", "model": "claude-x"}}},
            group=job_id,
            pin="judge",
            model="source",
            kind="task",
            producer="_judge",
            now=0.0,
        )
        assert lease(conn, config, "max") is None


def test_on_demand_task_success_is_collected_under_its_original_job(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", payload, 0.0)
        got = lease(conn, config, "max")
        report = CompletionReport(
            "succeeded",
            {"text": '{"ok":true}', "json": {"ok": True}},
            {},
            {"provider": "max-lane", "model": "claude-x"},
            None,
            1.0,
        )
        assert (
            complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0)
            == "succeeded"
        )
        assert get_job(conn, "pa", job_id)["state"] == "succeeded"


def test_on_demand_judge_records_scores_for_the_original_executor_labels(tmp_path):
    config, payload = setup_task(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        job_id, _ = submit_job(conn, config, "pa", payload, 0.0)
        source = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
        conn.execute("UPDATE jobs SET state='cancelled' WHERE id=?", (job_id,))
        insert_shadow_job(
            conn,
            source,
            {"prompt": "Judge.", "labels": {"A": {"provider": "ollama", "model": "model-a"}}},
            group=job_id,
            pin="judge",
            model="source",
            kind="task",
            producer="_judge",
            now=0.0,
        )
        got = lease(conn, config, "max")
        verdict = {"text": "", "json": {"scores": {"A": 5}, "best": "A"}}
        report = CompletionReport(
            "succeeded", verdict, {}, {"provider": "max-lane", "model": "claude-x"}, None, 1.0
        )
        assert (
            complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0)
            == "succeeded"
        )
        advance_shadow_groups(conn, config, 3.0)
        rows = read_judgements(conn, 0.0)
        assert [(r["provider"], r["mean_score"], r["best"]) for r in rows] == [("ollama", 5.0, 1)]
