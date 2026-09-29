"""Fold an attempt's outcome into the node's memory."""

from worker.node.node_memory import NodeMemory
from worker.node.pressure_samples import PRESSURE_SAMPLES

_MAX_BACKOFF_S = 7200.0
_FIRST_BACKOFF_S = 900.0
_RAN = ("succeeded", "preempted", "fenced", "drain_failed")


def settle_attempt(  # noqa: PLR0913, PLR0917
    memory: NodeMemory,
    model: str,
    was_resident: bool,
    outcome: str,
    code: str | None,
    now: float,
) -> bool:
    """Return True when the node should rest before asking for more work.

    A pressure release starts a backoff of 15 minutes, doubling per
    consecutive pressure release up to 2 hours; a success resets it. A model
    becomes the node's own only when the attempt ran on it: a failed attempt
    (``transport_error``, ``node_error``, an HTTP error) may never have loaded it.
    """
    if not was_resident and outcome in _RAN:
        memory.owned.add(model)
    if outcome == "drain_failed":
        memory.failed_model = model
    if outcome == "succeeded":
        memory.backoff_s = _FIRST_BACKOFF_S
        memory.backoff_until = 0.0
    if code == "memory_pressure" and outcome in ("preempted", "drain_failed"):
        memory.owned.add(model)
        memory.backoff_until = now + memory.backoff_s
        memory.backoff_s = min(memory.backoff_s * 2, _MAX_BACKOFF_S)
        memory.pressure_streak = max(memory.pressure_streak, PRESSURE_SAMPLES)
    return outcome == "failed" and code == "transport_error"
