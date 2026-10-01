import json

from tests.conftest import CONFIG, body, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.advance_shadow_groups import advance_shadow_groups
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.read_judgements import read_judgements
from worker.jobs.submit_job import submit_job

TARGETS = ["openrouter:vendor/model:free", "runner:bulk.agy"]


def make_config(tmp_path, max_pending=20):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {
        "bulk.agy": {"runner": "agy", "privacy": ["mail"]},
        "judge.codex": {"runner": "codex", "privacy": ["mail"]},
    }
    raw["queues"]["pa.bulk"]["shadow"] = {
        "rate": 1.0,
        "targets": TARGETS,
        "judge": "judge.codex",
        "max_pending": max_pending,
    }
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner", "openrouter"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def lease(conn, config, kind, now):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 0.0, 99.0, kind), now)


def finish(conn, config, got, provider, model, output, now):  # noqa: PLR0917
    executor = {"node": "node-a", "provider": provider, "model": model}
    report = CompletionReport("succeeded", output, {}, executor, None, 1.0)
    return complete_attempt(
        conn, config, got.attempt_id, got.generation, report, now, node="node-a"
    )


def answer(text):
    return {"text": text, "json": None}


def test_an_answer_is_shadowed_judged_blind_and_scored_per_executor(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    finish(
        conn, config, lease(conn, config, "inference", 1.0), "ollama", "model-a", answer("L"), 2.0
    )
    pins = sorted(r[0] for r in conn.execute("SELECT pin FROM jobs WHERE shadow_of IS NOT NULL"))
    assert pins == sorted(["original", *TARGETS])
    assert lease(conn, config, "inference", 3.0) is None
    remote = lease(conn, config, "openrouter", 3.0)
    assert remote.model == "vendor/model:free"
    finish(conn, config, remote, "openrouter", "vendor/model:free", answer("R"), 4.0)
    runner = lease(conn, config, "task", 5.0)
    assert runner.model == "bulk.agy"
    assert runner.input["prompt"] == "[user]\nhi"
    finish(conn, config, runner, "agy", "", answer("A"), 6.0)
    advance_shadow_groups(conn, config, 7.0)
    judge = lease(conn, config, "task", 8.0)
    assert judge.model == "judge.codex"
    assert conn.execute("SELECT max_attempts FROM jobs WHERE pin='judge'").fetchone()[0] == 2
    assert "=== Answer C ===" in judge.input["prompt"]
    labels = judge.input["labels"]
    best = next(k for k, v in labels.items() if v["provider"] == "agy")
    scores = {k: 5 if k == best else 2 for k in labels}
    verdict = {"text": "", "json": {"scores": scores, "best": best}}
    finish(conn, config, judge, "codex", "gpt", verdict, 9.0)
    advance_shadow_groups(conn, config, 10.0)
    judged = {(r["provider"], r["mean_score"], r["best"]) for r in read_judgements(conn, 0.0)}
    assert judged == {("agy", 5.0, 1), ("ollama", 2.0, 0), ("openrouter", 2.0, 0)}
    producers = [r[0] for r in conn.execute("SELECT producer FROM results")]
    assert producers.count("pa") == 1


def test_a_class_the_judge_may_not_see_is_never_shadowed(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(privacy="secret"), 0.0)
    finish(
        conn, config, lease(conn, config, "inference", 1.0), "ollama", "model-a", answer("L"), 2.0
    )
    assert conn.execute("SELECT count(*) FROM jobs").fetchone()[0] == 1


def test_no_group_opens_past_the_pending_cap(tmp_path):
    config = make_config(tmp_path, max_pending=2)
    conn = fresh_store(tmp_path / "state")
    for key in ("k1", "k2"):
        submit_job(conn, config, "pa", body(key=key), 0.0)
        got = lease(conn, config, "inference", 1.0)
        finish(conn, config, got, "ollama", "model-a", answer("L"), 2.0)
    assert conn.execute("SELECT count(DISTINCT shadow_of) FROM jobs").fetchone()[0] == 1


def test_a_group_with_one_answer_closes_without_a_judge(tmp_path):
    config = make_config(tmp_path)
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    finish(
        conn, config, lease(conn, config, "inference", 1.0), "ollama", "model-a", answer("L"), 2.0
    )
    conn.execute("UPDATE jobs SET state='failed' WHERE pin LIKE '%:%'")
    advance_shadow_groups(conn, config, 3.0)
    assert conn.execute("SELECT state FROM jobs WHERE pin='judge'").fetchone()[0] == "cancelled"
    assert lease(conn, config, "task", 4.0) is None


def test_a_pin_never_runs_under_a_fallback_or_a_class_its_profile_dropped(tmp_path):
    config = make_config(tmp_path)
    raw = json.loads((tmp_path / "config.json").read_text())
    raw["profiles"]["other.agy"] = {"runner": "agy", "privacy": ["mail"]}
    raw["queues"]["pa.bulk"]["profiles"] = ["bulk.agy", "other.agy"]
    raw["queues"]["pa.bulk"]["fallbacks"] = {"bulk.agy": ["other.agy"]}
    raw["profiles"]["bulk.agy"]["nodes"] = ["node-b"]
    (tmp_path / "config.json").write_text(json.dumps(raw))
    config = load_worker_config(tmp_path / "config.json")
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    finish(
        conn, config, lease(conn, config, "inference", 1.0), "ollama", "model-a", answer("L"), 2.0
    )
    finish(
        conn, config, lease(conn, config, "openrouter", 3.0), "openrouter", "m", answer("R"), 4.0
    )
    assert lease(conn, config, "task", 5.0) is None
    advance_shadow_groups(conn, config, 5.0 + 86400 * 2)
    raw["profiles"]["judge.codex"]["privacy"] = ["public"]
    (tmp_path / "config.json").write_text(json.dumps(raw))
    config = load_worker_config(tmp_path / "config.json")
    assert conn.execute("SELECT count(*) FROM jobs WHERE pin='judge'").fetchone()[0] == 1
    assert lease(conn, config, "task", 6.0 + 86400 * 2) is None
