"""Validate the shape of an inference ``input`` before it is stored."""

from typing import Any

from worker.jobs.api_error import ApiError


def check_inference_input(job_input: dict[str, Any]) -> None:
    """Raise ApiError(400, "bad_input") unless the node can run it as given.

    ``messages`` is a non-empty list of ``{role: str, content: str}``;
    ``options`` and ``schema`` are objects when present; ``format`` is ``"json"``.
    """
    messages = job_input.get("messages")
    if not isinstance(messages, list) or not messages:
        raise ApiError(400, "bad_input")
    for message in messages:
        if not (
            isinstance(message, dict)
            and isinstance(message.get("role"), str)
            and isinstance(message.get("content"), str)
        ):
            raise ApiError(400, "bad_input")
    for name in ("options", "schema"):
        if name in job_input and not isinstance(job_input[name], dict):
            raise ApiError(400, "bad_input")
    if "format" in job_input and job_input["format"] != "json":
        raise ApiError(400, "bad_input")
