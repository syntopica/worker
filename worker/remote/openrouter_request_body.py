"""The chat-completions body for one escalated inference job."""

from typing import Any

from worker.node.redact_credentials import redact_credentials


def openrouter_request_body(model: str, job_input: dict[str, Any], zdr: bool) -> dict[str, Any]:
    """Messages with credentials redacted; the job's schema becomes a strict ``json_schema`` format.

    ``zdr`` restricts routing to zero-data-retention endpoints.
    """
    messages = [
        {**m, "content": redact_credentials(str(m.get("content", "")))}
        for m in job_input["messages"]
    ]
    body: dict[str, Any] = {"model": model, "messages": messages}
    options = job_input.get("options") or {}
    if "temperature" in options:
        body["temperature"] = options["temperature"]
    schema = job_input.get("schema")
    if schema is not None:
        body["response_format"] = {
            "type": "json_schema",
            "json_schema": {"name": "output", "strict": True, "schema": schema},
        }
    if zdr:
        body["provider"] = {"zdr": True}
    return body
