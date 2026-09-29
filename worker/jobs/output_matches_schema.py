"""Check a node's output against the job's requested JSON schema."""

import json
import sqlite3

from jsonschema import Draft202012Validator, SchemaError
from referencing.exceptions import Unresolvable


def output_matches_schema(conn: sqlite3.Connection, job_id: str, output: object) -> bool:
    """A malformed or unresolvable schema counts as a violation, not a crash."""
    body = json.loads(
        conn.execute("SELECT body FROM p.inputs WHERE job_id=?", (job_id,)).fetchone()[0]
    )
    schema = body["input"].get("schema")
    if not isinstance(output, dict):
        return False
    if schema is None:
        return True
    try:
        Draft202012Validator.check_schema(schema)
        return not any(Draft202012Validator(schema).iter_errors(output.get("json")))
    except (SchemaError, Unresolvable):
        return False
