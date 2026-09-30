"""The answer inside an ``agy --output-format json`` envelope."""

import json


def agy_answer(stdout: str) -> str | None:
    """``structured_output`` first: with a schema, ``response`` is prose around it.

    Reading ``response`` first silently degraded every clips triage batch
    until 2026-08-02. Any malformed envelope is None.
    """
    try:
        envelope = json.loads(stdout)
    except ValueError:
        return None
    if not isinstance(envelope, dict):
        return None
    structured = envelope.get("structured_output")
    if isinstance(structured, dict | list):
        return json.dumps(structured)
    response = envelope.get("response")
    if isinstance(response, str):
        return response
    return json.dumps(response) if isinstance(response, dict | list) else None
