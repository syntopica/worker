"""Coerce a request field to a number, refusing anything else."""

import math
from typing import Any

from worker.jobs.api_error import ApiError


def coerce_number(value: Any, as_int: bool) -> int | float:
    """Return ``value`` as int or float; raise ApiError(400, "out_of_range")."""
    if isinstance(value, bool):
        raise ApiError(400, "out_of_range")
    try:
        number = int(value) if as_int else float(value)
    except (TypeError, ValueError, OverflowError) as error:
        raise ApiError(400, "out_of_range") from error
    if not math.isfinite(number):
        raise ApiError(400, "out_of_range")
    return number
