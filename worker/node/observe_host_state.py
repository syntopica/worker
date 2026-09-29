"""Fold one host sample into the node's memory."""

from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory

_PRESSED = ("warn", "critical")


def observe_host_state(memory: NodeMemory, state: HostState, now: float) -> None:
    """Track when pressure became normal and how long it has been pressured."""
    if state.pressure == "normal":
        memory.normal_since = memory.normal_since if memory.normal_since is not None else now
    else:
        memory.normal_since = None
    if state.pressure in _PRESSED:
        memory.pressure_streak += 1
    else:
        memory.pressure_streak = 0
        memory.unloaded = False  # a new pressure episode may unload again
