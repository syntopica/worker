"""Leave the tail of an unrecognised runner failure where an operator can read it."""

import sys


def report_runner_failure(runner: str, model: str) -> None:
    """Print executor metadata only; runner text can repeat private task content."""
    print(f"worker: runner_failed {runner}:{model}", file=sys.stderr)
