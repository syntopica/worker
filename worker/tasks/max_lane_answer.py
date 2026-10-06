"""The answer inside a ``max-lane-run`` envelope."""

import json


def max_lane_answer(stdout: str) -> str | None:
    """Free text as it is, a structured answer as JSON; any malformed envelope is None."""
    try:
        envelope = json.loads(stdout)
    except ValueError:
        return None
    answer = envelope.get("answer") if isinstance(envelope, dict) else None
    if isinstance(answer, str):
        return answer
    return json.dumps(answer) if isinstance(answer, dict | list) else None
