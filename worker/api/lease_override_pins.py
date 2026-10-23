"""Validate explicit permission to substitute a comparison's executor."""

from typing import Any

from worker.jobs.api_error import ApiError


def lease_override_pins(body: dict[str, Any], profile: str | None) -> bool:
    """Only an on-demand request may carry a true boolean override."""
    value = body.get("override_pins", False)
    if not isinstance(value, bool) or (value and profile is None):
        raise ApiError(400, "bad_request")
    return value
