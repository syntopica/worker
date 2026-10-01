"""The contract's output object for a runner's answer."""

from typing import Any

from worker.node.parse_json_answer import parse_json_answer


def task_output(answer: str) -> dict[str, Any]:
    """``json`` is the parsed answer when it is a JSON document (one fence allowed), else None."""
    return {"text": answer, "json": parse_json_answer(answer)}
