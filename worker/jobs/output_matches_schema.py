"""Check a node's output against the job's requested JSON schema."""

from typing import Any

from jsonschema import Draft202012Validator, SchemaError
from referencing.exceptions import Unresolvable


def output_matches_schema(job_input: dict[str, Any], output: object) -> bool:
    """A malformed or unresolvable schema counts as a violation, not a crash."""
    schema = job_input.get("schema")
    if not isinstance(output, dict):
        return False
    if schema is None:
        return True
    try:
        Draft202012Validator.check_schema(schema)
        return not any(Draft202012Validator(schema).iter_errors(output.get("json")))
    except (SchemaError, Unresolvable):
        return False
