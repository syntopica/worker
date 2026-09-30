"""Fold an attempt's outcome into the node's memory."""

from worker.node.node_memory import NodeMemory
from worker.node.pressure_samples import PRESSURE_SAMPLES

_MAX_BACKOFF_S = 7200.0
_TRANSPORT_BACKOFF_S = 30.0
_MAX_TRANSPORT_BACKOFF_S = 1800.0
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
    Consecutive transport failures hold the node back 30 s, doubling up to
    30 minutes, so a backend that lists models but cannot chat does not
    charge an attempt every rest.
    """
    if not was_resident and outcome in _RAN:
        memory.owned.add(model)
    if outcome == "drain_failed":
        memory.failed_model = model
    if outcome == "succeeded":
        memory.backoff_s = _FIRST_BACKOFF_S
        memory.backoff_until = 0.0
    transport = outcome == "failed" and code == "transport_error"
    memory.transport_failures = memory.transport_failures + 1 if transport else 0
    if transport:
        delay = _TRANSPORT_BACKOFF_S * 2 ** (memory.transport_failures - 1)
        memory.transport_until = now + min(delay, _MAX_TRANSPORT_BACKOFF_S)
    if code == "memory_pressure" and outcome in ("preempted", "drain_failed"):
        memory.owned.add(model)
        memory.backoff_until = now + memory.backoff_s
        memory.backoff_s = min(memory.backoff_s * 2, _MAX_BACKOFF_S)
        memory.pressure_streak = max(memory.pressure_streak, PRESSURE_SAMPLES)
    return transport
