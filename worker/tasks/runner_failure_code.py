"""Name why a runner failed from the end of its stderr, as a fixed code."""

from worker.tasks.runner_failure_patterns import RUNNER_FAILURE_PATTERNS

_TAIL_CHARS = 2048


def runner_failure_code(stderr: str) -> str:
    """The first matching code over the stderr tail, else ``runner_failed``."""
    tail = stderr[-_TAIL_CHARS:]
    for code, pattern in RUNNER_FAILURE_PATTERNS:
        if pattern.search(tail):
            return code
    return "runner_failed"
