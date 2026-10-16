"""How long a quota wall says it lasts, read from the runner's own wall message."""

import re

# agy since its weekly quotas (2026-10-16): "... Resets in 167h30m25s."
_RESETS_IN = re.compile(r"resets in (?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?", re.I)
# A reset further out than this is not trusted as a duration.
_MAX_S = 14 * 24 * 3600.0


def wall_resets_in_s(output: str) -> float | None:
    """Seconds until the wall resets, or None when the message names no duration."""
    for match in _RESETS_IN.finditer(output):
        hours, minutes, seconds = (int(part or 0) for part in match.groups())
        total = hours * 3600.0 + minutes * 60.0 + seconds
        if 0 < total <= _MAX_S:
            return total
    return None
