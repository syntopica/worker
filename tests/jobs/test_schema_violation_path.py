import json

from tests.conftest import body
from tests.jobs.test_complete_attempt import SCHEMA, report, run
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.schema_violation_path import schema_violation_path
from worker.jobs.submit_job import submit_job


def test_the_path_names_the_schema_rule_never_the_output():
    schema = {"type": "object", "additionalProperties": {"type": "string"}}
    path = schema_violation_path({"schema": schema}, {"json": {"secret-key": 5}})
    assert path == "/additionalProperties/type"
    assert "secret" not in path


def test_no_path_when_the_output_matches_or_the_schema_is_unusable():
    assert schema_violation_path({"schema": SCHEMA}, {"json": {"label": "x"}}) is None
    assert schema_violation_path({"schema": {"$ref": "#/nope"}}, {"json": {}}) is None
    assert schema_violation_path({"schema": SCHEMA}, "text") is None


def test_the_final_failed_result_carries_the_schema_path(conn, config):
    job_input = {"messages": [{"role": "user", "content": "hi"}], "schema": SCHEMA}
    submit_job(conn, config, "pa", body(input=job_input, max_attempts=1), 0.0)
    lease = run(conn, config)
    bad = report("succeeded", output={"text": "", "json": {"label": 3}})
    assert complete_attempt(conn, config, lease.attempt_id, lease.generation, bad, 2.0) == "failed"
    detail = json.loads(conn.execute("SELECT detail FROM results").fetchone()[0])
    assert detail == {"error": "schema_violation", "schema_path": "/properties/label/type"}
