"""Log one residency change with its UTC time, for the 1c reload count."""

import sys

from worker.node.utc_stamp import utc_stamp


def log_residency_change(kind: str, model: str) -> None:
    """``worker: residency load|unload model=<name> at=<UTC>`` on stderr.

    The rolling hour in the node report cannot be summed afterwards; these
    lines can be matched against the queue's history to count the loads that
    happened while work was queued.
    """
    print(f"worker: residency {kind} model={model} at={utc_stamp()}", file=sys.stderr)
