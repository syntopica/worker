"""Check a node's output against the job's requested JSON schema."""

import json
import sqlite3

from jsonschema import Draft202012Validator, SchemaError


def output_matches_schema(
    conn: sqlite3.Connection, job_id: str, output: dict[str, object] | None
) -> bool:
    """A malformed schema counts as a violation, not a crash."""
    body = json.loads(
        conn.execute("SELECT body FROM p.inputs WHERE job_id=?", (job_id,)).fetchone()[0]
    )
    schema = body["input"].get("schema")
    if not schema:
        return output is not None
    if output is None:
        return False
    try:
        Draft202012Validator.check_schema(schema)
        return not any(Draft202012Validator(schema).iter_errors(output.get("json")))
    except SchemaError:
        return False
