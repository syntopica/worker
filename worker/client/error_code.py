"""Read the error code out of an HTTP error body."""

import json


def error_code(raw: bytes) -> str:
    """The ``error`` field of an object body, or ``unknown`` for anything else."""
    try:
        parsed = json.loads(raw)
    except ValueError:
        return "unknown"
    if not isinstance(parsed, dict):
        return "unknown"
    return str(parsed.get("error", "unknown"))
