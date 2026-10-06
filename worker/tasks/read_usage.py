"""Read a finished runner's token usage from wherever its runner reports it."""

from typing import Any

from worker.tasks.envelope_usage import envelope_usage


def read_usage(runner: str, stdout: str) -> dict[str, Any]:
    """The agy, cursor and max-lane envelopes carry ``usage``; codex reports none.

    Field names are pinned by captured envelopes in ``tests/tasks/fixtures``.
    Codex, run without ``--json``, prints only a locale-formatted total on its
    human stderr, which splits neither direction, so it stays unmeasured.
    """
    if runner in {"agy", "max-lane"}:
        return envelope_usage(stdout, "input_tokens", "output_tokens")
    if runner == "cursor":
        return envelope_usage(stdout, "inputTokens", "outputTokens")
    return {}
