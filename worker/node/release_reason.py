"""Why an attempt must be released now, if it must (spec 7)."""

from worker.node.host_state import HostState

USER_RETURN_S = 10.0


def release_reason(state: HostState, run_when: str) -> str | None:
    """Input releases idle-only work; battery and pressure release everything."""
    if state.idle_s is None or state.on_ac is None or state.pressure == "unknown":
        return "host_state_unreadable"
    if not state.on_ac:
        return "on_battery"
    if state.pressure != "normal":
        return "memory_pressure"
    if run_when == "idle" and state.idle_s < USER_RETURN_S:
        return "user_active"
    return None
