"""Why the node may not take any work right now, if it may not."""

from worker.node.host_state import HostState
from worker.node.release_reason import release_reason


def admission_block(
    state: HostState, normal_since: float | None, now: float, recovery_s: float = 120.0
) -> str | None:
    """Release conditions, plus the pressure recovery window before readmission."""
    reason = release_reason(state, "active_ok")
    if reason is not None:
        return reason
    if normal_since is None or now - normal_since < recovery_s:
        return "pressure_recovering"
    return None
