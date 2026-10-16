"""Leave the tail of an unrecognised runner failure where an operator can read it."""

import sys

from worker.node.redact_credentials import redact_credentials

_TAIL = 300


def report_runner_failure(runner: str, model: str, text: str) -> None:
    """Print the runner's own failure text, redacted and cut to its tail.

    The scratch holding stderr is removed with the attempt, so an unmatched
    failure otherwise left only ``runner_failed`` behind: on 2026-10-16 that
    hid, for a whole day, that agy no longer offered the profile's model.
    """
    tail = " ".join(redact_credentials(text).split())[-_TAIL:]
    print(f"worker: runner_failed {runner}:{model}: {tail}", file=sys.stderr)
