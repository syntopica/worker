"""The reset a node reported for a quota wall, if it is a duration worth trusting."""

import math
from typing import Any

# Same bound the node reads with: a weekly quota, with room to spare.
_MAX_S = 14 * 24 * 3600.0


def trusted_wall_reset(usage: dict[str, Any]) -> float | None:
    """``wall_resets_in_s`` when it is a finite number of seconds in (0, 14 days]."""
    value = usage.get("wall_resets_in_s")
    if isinstance(value, bool) or not isinstance(value, int | float):
        return None
    seconds = float(value)
    return seconds if math.isfinite(seconds) and 0 < seconds <= _MAX_S else None
