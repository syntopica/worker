"""Hand a leased job back when the loop running it is stopped."""

import sys
from typing import Any

from worker.node.attempt_report import attempt_report


def report_shutdown(
    link: Any, attempt: str, generation: int, executor: dict[str, str], wall_s: float
) -> None:
    """Best effort: complete as ``preempted``/``node_shutdown``, which charges nothing.

    If the coordinator cannot be reached the lease simply expires, as before.
    """
    report = attempt_report("preempted", executor, wall_s, error_code="node_shutdown")
    try:
        link.complete(attempt, generation, report)
    except Exception as error:
        print(f"worker: shutdown report failed: {type(error).__name__}", file=sys.stderr)
