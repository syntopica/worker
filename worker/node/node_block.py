"""Why the node may not ask for work in this iteration, if it may not."""

from worker.node.admission_block import admission_block
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory


def node_block(memory: NodeMemory, state: HostState, now: float, backend_down: bool) -> str | None:
    """A failed drain first, then host state, the backend, and the pressure backoff."""
    if memory.failed_model is not None:
        return "drain_failed"
    reason = admission_block(state, memory.normal_since, now)
    if reason is not None:
        return reason
    if backend_down:
        return "backend_down"
    if now < memory.backoff_until:
        return "pressure_backoff"
    return None
