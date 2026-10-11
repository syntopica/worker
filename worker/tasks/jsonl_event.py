"""One line of a runner's JSON Lines stream, as an event object."""

import json
from typing import Any


def jsonl_event(line: str) -> dict[str, Any] | None:
    """The line's JSON object, or None for a blank, non-JSON or non-object line.

    Runners can print plain warnings among their events, so a line that does
    not parse is skipped rather than failing the whole stream.
    """
    if not line.startswith("{"):
        return None
    try:
        event = json.loads(line)
    except ValueError:
        return None
    return event if isinstance(event, dict) else None
