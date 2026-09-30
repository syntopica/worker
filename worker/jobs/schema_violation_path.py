"""Where an output first breaks its schema, named by the schema's own path."""

from typing import Any

from jsonschema import Draft202012Validator, SchemaError
from referencing.exceptions import Unresolvable


def schema_violation_path(job_input: dict[str, Any], output: object) -> str | None:
    """A JSON pointer into the schema (``/properties/id/type``), never into the output.

    The schema is the producer's own, so its path carries no content; the
    output's path could (an ``additionalProperties`` key is data). None when
    the output is not an object, the schema is unusable, or nothing fails.
    """
    schema = job_input.get("schema", job_input.get("output_schema"))
    if schema is None or not isinstance(output, dict):
        return None
    try:
        Draft202012Validator.check_schema(schema)
        error = next(Draft202012Validator(schema).iter_errors(output.get("json")), None)
    except (SchemaError, Unresolvable):
        return None
    if error is None:
        return None
    return "/" + "/".join(str(part) for part in error.absolute_schema_path)
