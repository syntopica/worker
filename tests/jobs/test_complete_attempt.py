import json
import socket
import urllib.request

import pytest

from tests.conftest import body
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job

SCHEMA = {"type": "object", "required": ["label"], "properties": {"label": {"type": "string"}}}

MSG = [{"role": "user", "content": "hi"}]


def run(conn, config, now=1.0):
    return lease_job(conn, config, LeaseRequest("node-a", None, False, 44.0, 9999.0), now)


def report(outcome, output=None, code=None):
    return CompletionReport(
        outcome,
        output,
        {"tokens_in": 3, "tokens_out": 2},
        {"node": "node-a", "provider": "ollama", "model": "model-a"},
        code,
        5.0,
    )


def test_valid_output_becomes_a_result(conn, config):
    submit_job(conn, config, "pa", body(input={"messages": MSG, "schema": SCHEMA}), 0.0)
    lease = run(conn, config)
    state = complete_attempt(
        conn,
        config,
        lease.attempt_id,
        lease.generation,
        report("succeeded", {"text": "{}", "json": {"label": "x"}}),
        2.0,
    )
    assert state == "succeeded"
    result = conn.execute("SELECT result_id, control FROM results").fetchone()
    stored = conn.execute(
        "SELECT body FROM p.outputs WHERE result_id=?", (result["result_id"],)
    ).fetchone()[0]
    assert (result["control"], json.loads(stored)["json"]) == (None, {"label": "x"})


def test_schema_violation_is_a_failed_attempt_with_backoff(conn, config):
    submit_job(conn, config, "pa", body(input={"messages": MSG, "schema": SCHEMA}), 0.0)
    lease = run(conn, config)
    state = complete_attempt(
        conn,
        config,
        lease.attempt_id,
        lease.generation,
        report("succeeded", {"text": "{}", "json": {}}),
        2.0,
    )
    job = conn.execute("SELECT attempts, not_before, error FROM jobs").fetchone()
    assert (state, job["attempts"], job["not_before"], job["error"]) == (
        "queued",
        1,
        32.0,
        "schema_violation",
    )


def test_preemption_does_not_spend_an_attempt_and_the_third_asks_for_a_split(conn, config):
    submit_job(conn, config, "pa", body(), 0.0)
    states = []
    for i in range(3):
        lease = run(conn, config, now=10.0 * (i + 1))
        states.append(
            complete_attempt(
                conn,
                config,
                lease.attempt_id,
                lease.generation,
                report("preempted", code="user_active"),
                10.0 * (i + 1) + 1,
            )
        )
    assert states == ["queued", "queued", "split_requested"]
    assert conn.execute("SELECT attempts FROM jobs").fetchone()[0] == 0
    assert conn.execute("SELECT control FROM results").fetchone()[0] == "split_requested"


def test_parked_runs_double_the_idle_needed_then_exhaust(conn, config):
    submit_job(conn, config, "pa", body(), 0.0)
    conn.execute("UPDATE jobs SET parked=1, parked_min_idle_s=100, preemptions=3")
    states = []
    for i in range(3):
        lease = run(conn, config, now=100.0 * (i + 1))
        states.append(
            complete_attempt(
                conn,
                config,
                lease.attempt_id,
                lease.generation,
                report("preempted", code="user_active"),
                100.0 * (i + 1) + 1,
            )
        )
    row = conn.execute("SELECT state, error, parked_min_idle_s FROM jobs").fetchone()
    assert states == ["queued", "queued", "failed"]
    assert (row["error"], row["parked_min_idle_s"]) == ("preemption_exhausted", 400.0)


def test_malformed_schema_is_a_schema_violation_not_a_crash(conn, config):
    submit_job(conn, config, "pa", body(input={"messages": MSG, "schema": {"type": 12}}), 0.0)
    lease = run(conn, config)
    state = complete_attempt(
        conn,
        config,
        lease.attempt_id,
        lease.generation,
        report("succeeded", {"text": "{}", "json": {}}),
        2.0,
    )
    assert (state, conn.execute("SELECT error FROM jobs").fetchone()[0]) == (
        "queued",
        "schema_violation",
    )


def _violation(conn, config, schema, output=None):
    submit_job(conn, config, "pa", body(input={"messages": MSG, "schema": schema}), 0.0)
    lease = run(conn, config)
    out = {"text": "{}", "json": {}} if output is None else output
    state = complete_attempt(
        conn, config, lease.attempt_id, lease.generation, report("succeeded", out), 2.0
    )
    return state, conn.execute("SELECT error FROM jobs").fetchone()[0]


def test_dangling_local_ref_is_a_schema_violation(conn, config):
    assert _violation(conn, config, {"$ref": "#/nope"}) == ("queued", "schema_violation")


def test_remote_ref_is_a_schema_violation_without_network(conn, config, monkeypatch):

    def refuse(*_a, **_k):
        raise AssertionError("network retrieval attempted")

    monkeypatch.setattr(socket.socket, "connect", refuse)
    monkeypatch.setattr(urllib.request, "urlopen", refuse)
    assert _violation(conn, config, {"$ref": "http://x.invalid/s"}) == (
        "queued",
        "schema_violation",
    )


def test_a_non_object_schema_is_refused_at_submit(conn, config):
    with pytest.raises(ApiError) as error:
        _violation(conn, config, False)
    assert error.value.code == "bad_input"


def test_non_dict_output_fails_the_check(conn, config):
    assert _violation(conn, config, SCHEMA, output=["x"]) == ("queued", "schema_violation")
    attempt = conn.execute("SELECT outcome, error FROM attempts").fetchone()
    assert (attempt["outcome"], attempt["error"]) == ("schema_violation", "schema_violation")
