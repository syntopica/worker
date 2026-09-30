"""The answer inside a ``cursor-agent --output-format json`` envelope."""

import json

from worker.tasks.last_json_object import last_json_object


def cursor_answer(stdout: str) -> str | None:
    """``result`` unless ``is_error``; its last JSON object when it narrates first.

    Cursor has no output schema, and a tool-using model says something before
    the object it was asked for, so the final balanced object is the answer.
    """
    try:
        envelope = json.loads(stdout)
    except ValueError:
        return None
    if not isinstance(envelope, dict) or envelope.get("is_error") is True:
        return None
    result = envelope.get("result")
    if not isinstance(result, str):
        return None
    return last_json_object(result) or result
