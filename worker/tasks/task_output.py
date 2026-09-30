"""The contract's output object for a runner's answer."""

import json
from typing import Any


def task_output(answer: str) -> dict[str, Any]:
    """``json`` is the parsed answer when it is a JSON document, else None."""
    try:
        parsed: Any = json.loads(answer)
    except ValueError:
        parsed = None
    return {"text": answer, "json": parsed}
