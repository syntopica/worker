"""Name the executor of a result for a script's log line."""

from typing import Any


def format_executor(executor: dict[str, Any] | None) -> str:
    """``<provider> <model>``, each ``unknown`` when the node did not report it.

    Model ids carry ``/`` and ``:``, so the two are separated by a space.
    """
    executor = executor or {}
    provider = executor.get("provider") or "unknown"
    model = executor.get("model") or "unknown"
    return f"{provider} {model}"
