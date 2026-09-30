"""Count model loads and unloads the node sees, for the reload criterion of 1c."""

from worker.node.node_memory import NodeMemory

_WINDOW_S = 3600.0


def track_residency(memory: NodeMemory, resident: list[str], now: float) -> dict[str, int]:
    """Record residency changes since the last sample; return the last hour's counts.

    A change is counted whatever caused it (a lease, a pressure unload, a
    keep-alive expiry or another application), because the acceptance
    criterion is about reloads, not about who caused them.
    """
    current = set(resident)
    if memory.last_resident is not None:
        memory.residency_events.extend((now, "load") for _ in current - memory.last_resident)
        memory.residency_events.extend((now, "unload") for _ in memory.last_resident - current)
    memory.last_resident = current
    memory.residency_events[:] = [e for e in memory.residency_events if now - e[0] < _WINDOW_S]
    loads = sum(1 for _, kind in memory.residency_events if kind == "load")
    return {"loads_1h": loads, "unloads_1h": len(memory.residency_events) - loads}
