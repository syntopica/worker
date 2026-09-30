import json

import pytest

from tests.conftest import CONFIG, fresh_store
from worker.config.load_worker_config import load_worker_config
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job

SCHEMA = {"type": "object", "required": ["label"], "properties": {"label": {"type": "string"}}}


@pytest.fixture
def tconfig(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.refine"] = {"run_when": "idle", "profiles": ["pa.codex", "pa.grade"]}
    raw["producers"]["pa"].append("pa.refine")
    raw["profiles"] = {
        "pa.codex": {"runner": "codex", "model": "m-1", "reasoning": "low"},
        "pa.grade": {"runner": "codex", "input_root": "store", "nodes": ["node-a"]},
        "pa.agy": {"runner": "agy"},
    }
    raw["runners"] = {"codex": {"cooldown_s": 600}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


@pytest.fixture
def tconn(tmp_path):
    return fresh_store(tmp_path / "state")


def task(key="t1", profile="pa.codex", runner="codex", privacy="internal", **extra):
    task_input = {"runner": runner, "profile": profile, "prompt": "classify", **extra}
    return json.dumps(
        {
            "contract": 1,
            "kind": "task",
            "queue": "pa.refine",
            "idempotency_key": key,
            "privacy": privacy,
            "max_attempts": 2,
            "input": task_input,
        }
    ).encode()


def lease(conn, config, kind="task", node="node-a", now=1.0):
    return lease_job(conn, config, LeaseRequest(node, None, False, 0.0, 9999.0, kind), now)


def report(outcome, output=None, code=None):
    executor = {"node": "node-a", "provider": "codex", "model": "m-1"}
    return CompletionReport(outcome, output, {}, executor, code, 3.0)


def test_a_granted_task_is_stored_under_its_profile_and_leased_only_as_a_task(tconn, tconfig):
    job_id, created = submit_job(tconn, tconfig, "pa", task(), 0.0)
    row = tconn.execute("SELECT kind, model FROM jobs WHERE id=?", (job_id,)).fetchone()
    assert created and (row["kind"], row["model"]) == ("task", "pa.codex")
    assert lease(tconn, tconfig, kind="inference") is None
    granted = lease(tconn, tconfig)
    assert granted is not None
    assert (granted.kind, granted.model, granted.input["prompt"]) == (
        "task",
        "pa.codex",
        "classify",
    )


@pytest.mark.parametrize(
    ("raw", "status", "code"),
    [
        (task(profile="pa.agy", runner="agy"), 403, "profile_not_granted"),
        (task(profile="nope"), 403, "profile_not_granted"),
        (task(runner="cursor"), 400, "runner_mismatch"),
        (task(privacy="mail"), 403, "privacy_not_allowed"),
        (task(inputs=["a.md"]), 400, "inputs_not_allowed"),
        (task(profile="pa.grade", inputs=["../secret"]), 400, "bad_input"),
        (task(profile="pa.grade", inputs=["/etc/passwd"]), 400, "bad_input"),
        (task(output_schema="x"), 400, "bad_input"),
        (task(extra_field=1), 400, "bad_input"),
    ],
)
def test_task_refusals(tconn, tconfig, raw, status, code):
    with pytest.raises(ApiError) as error:
        submit_job(tconn, tconfig, "pa", raw, 0.0)
    assert (error.value.status, error.value.code) == (status, code)


def test_personal_classes_never_reach_a_runner_node(tconn, tconfig):
    tconn.execute("UPDATE jobs SET privacy='mail'")
    submit_job(tconn, tconfig, "pa", task(), 0.0)
    tconn.execute("UPDATE jobs SET privacy='mail'")
    assert lease(tconn, tconfig) is None


def test_a_profile_node_allowlist_is_enforced(tconn, tconfig):
    submit_job(tconn, tconfig, "pa", task(profile="pa.grade", inputs=["a.md"]), 0.0)
    assert lease(tconn, tconfig, node="node-g") is None
    assert lease(tconn, tconfig, node="node-a") is not None


def test_a_quota_wall_rests_the_runner_and_requeues_without_charging(tconn, tconfig):
    submit_job(tconn, tconfig, "pa", task("t1"), 0.0)
    submit_job(tconn, tconfig, "pa", task("t2"), 0.0)
    first = lease(tconn, tconfig)
    state = complete_attempt(
        tconn, tconfig, first.attempt_id, first.generation, report("failed", code="quota_wall"), 2.0
    )
    assert state == "queued"
    row = tconn.execute(
        "SELECT attempts, not_before FROM jobs WHERE id=?", (first.job_id,)
    ).fetchone()
    assert (row["attempts"], row["not_before"]) == (0, 602.0)
    assert lease(tconn, tconfig, now=3.0) is None  # the other codex task waits too
    assert lease(tconn, tconfig, now=603.0) is not None


def test_task_output_is_checked_against_output_schema(tconn, tconfig):
    submit_job(tconn, tconfig, "pa", task(output_schema=SCHEMA), 0.0)
    first = lease(tconn, tconfig)
    bad = report("succeeded", {"text": "{}", "json": {}})
    assert (
        complete_attempt(tconn, tconfig, first.attempt_id, first.generation, bad, 2.0) == "queued"
    )
    second = lease(tconn, tconfig, now=100.0)
    good = report("succeeded", {"text": "", "json": {"label": "x"}})
    assert (
        complete_attempt(tconn, tconfig, second.attempt_id, second.generation, good, 101.0)
        == "succeeded"
    )


def test_input_root_resolves_against_the_config_file(tconfig, tmp_path):
    assert tconfig.profiles["pa.grade"].input_root == (tmp_path / "store").resolve()
    assert tconfig.profiles["pa.codex"].privacy == frozenset({"public", "internal"})
