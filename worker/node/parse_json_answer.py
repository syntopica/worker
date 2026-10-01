"""Parse a model's answer as JSON, tolerating one surrounding code fence."""

import json
from typing import Any


def parse_json_answer(text: str) -> Any:
    """The parsed document, or None when the answer is not JSON.

    Some free models and agy wrap JSON in a ```json fence even when asked for
    JSON only; a single surrounding fence is removed before parsing.
    """
    body = text.strip()
    if body.startswith("```") and body.endswith("```"):
        body = body.strip("`").removeprefix("json").strip()
    try:
        return json.loads(body)
    except ValueError:
        return None
