"""Token counts summed over the ``turn.completed`` events of ``codex exec --json``."""

from typing import Any

from worker.tasks.jsonl_event import jsonl_event


def codex_usage(stdout: str) -> dict[str, Any]:
    """``tokens_in`` and ``tokens_out`` over every completed turn, else nothing.

    ``input_tokens`` already counts the cached part and ``output_tokens`` the
    reasoning part, so the two are taken as they are. A run in which no turn
    completed reports no usage rather than zero; a non-integer count is left
    out rather than guessed. Codex prints no price, so ``cost_usd`` stays unset.
    """
    totals: dict[str, int] = {}
    for line in stdout.splitlines():
        event = jsonl_event(line)
        if event is None or event.get("type") != "turn.completed":
            continue
        usage = event.get("usage")
        if not isinstance(usage, dict):
            continue
        for field, key in (("tokens_in", "input_tokens"), ("tokens_out", "output_tokens")):
            count = usage.get(key)
            if isinstance(count, int) and not isinstance(count, bool):
                totals[field] = totals.get(field, 0) + count
    return dict(totals)
