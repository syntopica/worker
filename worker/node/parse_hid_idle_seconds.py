"""Seconds since the last keyboard or pointer input, from ``ioreg -c IOHIDSystem``."""

import re

_IDLE = re.compile(r'"HIDIdleTime" = (\d+)')


def parse_hid_idle_seconds(text: str) -> float | None:
    """HIDIdleTime is in nanoseconds and covers every user at the console."""
    match = _IDLE.search(text)
    return int(match.group(1)) / 1e9 if match else None
