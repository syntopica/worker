"""How long a quota wall says it lasts, read from the runner's own wall message."""

import re

from worker.tasks.quota_wall_patterns import QUOTA_WALL_PATTERNS

# agy since its weekly quotas (2026-10-16): "... Resets in 167h30m25s."
_RESETS_IN = re.compile(r"resets in (?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?", re.I)
# A reset further out than this is not trusted as a duration.
_MAX_S = 14 * 24 * 3600.0
# The duration must follow the wall sentence itself: elsewhere in the output it
# may be the model's prose over untrusted input.
_AFTER_WALL = 200


def wall_resets_in_s(runner: str, output: str) -> float | None:
    """Seconds until the wall resets, read right after the runner's wall sentence."""
    pattern = QUOTA_WALL_PATTERNS.get(runner)
    wall = pattern.search(output) if pattern is not None else None
    if wall is None:
        return None
    match = _RESETS_IN.search(output, wall.end(), wall.end() + _AFTER_WALL)
    if match is None:
        return None
    hours, minutes, seconds = (int(part or 0) for part in match.groups())
    total = hours * 3600.0 + minutes * 60.0 + seconds
    return total if 0 < total <= _MAX_S else None
