"""Read a finished runner's token usage from wherever its runner reports it."""

from typing import Any

from worker.tasks.codex_usage import codex_usage
from worker.tasks.envelope_usage import envelope_usage


def read_usage(runner: str, stdout: str) -> dict[str, Any]:
    """The agy, cursor and max-lane envelopes carry ``usage``; codex its events.

    Field names are pinned by captured output in ``tests/tasks/fixtures``.
    """
    if runner in {"agy", "max-lane"}:
        return envelope_usage(stdout, "input_tokens", "output_tokens")
    if runner == "codex":
        return codex_usage(stdout)
    if runner == "cursor":
        return envelope_usage(stdout, "inputTokens", "outputTokens")
    return {}
