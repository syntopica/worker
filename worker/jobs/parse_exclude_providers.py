"""The executor providers a submit body excludes."""

from typing import Any

from worker.jobs.api_error import ApiError


def parse_exclude_providers(requirements: Any) -> tuple[str, ...]:
    """Absent or empty excludes nothing; anything but a list of non-empty strings is refused.

    Raise ApiError(400, ``bad_requirements``) on a malformed value. Order is kept
    and repeats dropped.
    """
    if not isinstance(requirements, dict) or "exclude_providers" not in requirements:
        return ()
    value = requirements["exclude_providers"]
    if not isinstance(value, list) or not all(isinstance(p, str) and p for p in value):
        raise ApiError(400, "bad_requirements")
    return tuple(dict.fromkeys(value))
