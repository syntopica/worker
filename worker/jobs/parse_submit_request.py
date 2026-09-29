"""Validate a decoded submit body against contract v1."""

from typing import Any

from worker.jobs.api_error import ApiError
from worker.jobs.states import PRIVACY_CLASSES
from worker.jobs.submit_request import SubmitRequest

_MAX_KEY_LENGTH = 256
_MAX_PRIORITY = 100
_MAX_ATTEMPTS = 10


def parse_submit_request(body: dict[str, Any]) -> SubmitRequest:
    """Return the request or raise ApiError(400) naming the first problem."""
    if body.get("contract") != 1:
        raise ApiError(400, "unsupported_contract")
    if body.get("kind") != "inference":
        raise ApiError(400, "unsupported_kind")
    if body.get("privacy") not in PRIVACY_CLASSES:
        raise ApiError(400, "unknown_privacy")
    models = (body.get("requirements") or {}).get("models")
    if not isinstance(models, list) or not models:
        raise ApiError(400, "missing_models")
    key = body.get("idempotency_key")
    if not isinstance(key, str) or not key or len(key) > _MAX_KEY_LENGTH:
        raise ApiError(400, "bad_idempotency_key")
    priority = int(body.get("priority", 50))
    attempts = int(body.get("max_attempts", 3))
    if not 0 <= priority <= _MAX_PRIORITY or not 1 <= attempts <= _MAX_ATTEMPTS:
        raise ApiError(400, "out_of_range")
    if not isinstance(body.get("input"), dict):
        raise ApiError(400, "missing_input")
    deadline = body.get("deadline")
    return SubmitRequest(
        queue=str(body.get("queue", "")),
        idempotency_key=key,
        priority=priority,
        privacy=body["privacy"],
        max_attempts=attempts,
        deadline=float(deadline) if deadline is not None else None,
        models=tuple(str(m) for m in models),
        input=body["input"],
        parent_id=body.get("parent_id"),
    )
