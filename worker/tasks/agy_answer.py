"""The answer inside an agy final result or legacy ``--output-format json`` envelope."""

import json

from worker.tasks.agy_final_result import agy_final_result


def agy_answer(stdout: str) -> str | None:
    """``structured_output`` first: with a schema, ``response`` is prose around it.

    Reading ``response`` first silently degraded every clips triage batch
    until 2026-08-02. A stream without a final ``SUCCESS`` result, or any
    malformed envelope, is None.
    """
    final = agy_final_result(stdout)
    if final is None:
        return None
    envelope, streamed = final
    if streamed and envelope.get("status") != "SUCCESS":
        return None
    structured = envelope.get("structured_output")
    if isinstance(structured, dict | list):
        return json.dumps(structured)
    response = envelope.get("response")
    if isinstance(response, str):
        return response
    return json.dumps(response) if isinstance(response, dict | list) else None
