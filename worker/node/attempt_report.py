"""The completion report body for one attempt."""

from typing import Any


def attempt_report(  # noqa: PLR0913
    outcome: str,
    executor: dict[str, str],
    wall_s: float,
    *,
    output: Any = None,
    usage: dict[str, Any] | None = None,
    error_code: str | None = None,
) -> dict[str, Any]:
    """Report fields of POST /v1/attempts/{id}/complete, without the generation."""
    return {
        "outcome": outcome,
        "output": output,
        "usage": usage or {},
        "executor": executor,
        "error_code": error_code,
        "wall_s": wall_s,
    }
