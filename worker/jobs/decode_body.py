"""Decode a raw submit body into a JSON object."""

import json
from typing import Any

from worker.jobs.api_error import ApiError


def decode_body(raw: bytes) -> dict[str, Any]:
    """Return the JSON object or raise ApiError(400, "bad_json")."""
    try:
        body = json.loads(raw)
    except (ValueError, RecursionError) as error:
        raise ApiError(400, "bad_json") from error
    if not isinstance(body, dict):
        raise ApiError(400, "bad_json")
    return body
