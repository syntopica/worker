"""The answer inside a ``claude -p --output-format json`` envelope."""

import json


def claude_answer(stdout: str) -> str | None:
    """``structured_output`` as JSON, else ``result``; None for an error envelope."""
    try:
        envelope = json.loads(stdout)
    except ValueError:
        return None
    if not isinstance(envelope, dict) or envelope.get("is_error") is True:
        return None
    structured = envelope.get("structured_output")
    if isinstance(structured, dict | list):
        return json.dumps(structured)
    result = envelope.get("result")
    return result if isinstance(result, str) else None
