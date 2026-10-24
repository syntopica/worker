"""The terminal result of an agy run, from its event stream or a legacy envelope."""

import json
from typing import Any


def agy_final_result(stdout: str) -> tuple[dict[str, Any], bool] | None:
    """``(result, streamed)``, or None when stdout holds no complete final result.

    A ``stream-json`` run ends with one ``{"event":"result","result":{...}}``
    line; step updates before it are never an answer, and a truncated stream
    has no such last line. A single JSON envelope without an ``event`` key is
    the older ``--output-format json`` shape, still read for rolling upgrades.
    """
    text = stdout.strip()
    if not text:
        return None
    try:
        whole = json.loads(text)
    except ValueError:
        whole = None
    if isinstance(whole, dict) and "event" not in whole:
        return whole, False
    try:
        last = json.loads(text.split("\n")[-1])
    except ValueError:
        return None
    if not isinstance(last, dict) or last.get("event") != "result":
        return None
    result = last.get("result")
    return (result, True) if isinstance(result, dict) else None
