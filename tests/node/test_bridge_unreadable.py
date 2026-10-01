from worker.node.bridge_unreadable import BridgeUnreadable
from worker.node.host_state import HostState


def test_a_failed_idle_or_pressure_read_reuses_one_at_most_ten_seconds_old():
    samples = [HostState(30, True, "normal"), HostState(None, True, "unknown")]
    now = [0.0]
    bridge = BridgeUnreadable(
        lambda: samples.pop(0) if len(samples) > 1 else samples[0], lambda: now[0]
    )
    assert bridge() == HostState(30, True, "normal")
    now[0] = 8.0
    assert bridge() == HostState(30, True, "normal")
    now[0] = 12.0
    assert bridge() == HostState(None, True, "unknown")


def test_without_any_reading_nothing_is_invented():
    bridge = BridgeUnreadable(lambda: HostState(None, True, "unknown"), lambda: 0.0)
    assert bridge() == HostState(None, True, "unknown")
