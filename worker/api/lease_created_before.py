"""Validate the optional creation-time boundary for an explicit drain."""

import math
from typing import Any

from worker.jobs.api_error import ApiError


def lease_created_before(body: dict[str, Any], profile: str | None) -> float | None:
    """Return epoch seconds; a boundary is valid only with an explicit profile."""
    value = body.get("created_before")
    if value is None:
        return None
    if (
        profile is None
        or isinstance(value, bool)
        or not isinstance(value, (int, float))
        or value < 0
    ):
        raise ApiError(400, "bad_request")
    try:
        cutoff = float(value)
    except OverflowError:
        raise ApiError(400, "bad_request") from None
    if not math.isfinite(cutoff):
        raise ApiError(400, "bad_request")
    return cutoff
