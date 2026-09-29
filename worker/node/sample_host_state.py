"""Read idle time, power source and memory pressure."""

from collections.abc import Callable

from worker.node.host_state import HostState
from worker.node.parse_hid_idle_seconds import parse_hid_idle_seconds
from worker.node.parse_on_ac import parse_on_ac
from worker.node.parse_pressure import parse_pressure
from worker.node.run_command import run_command


def sample_host_state(run: Callable[[list[str]], str | None] = run_command) -> HostState:
    """Three short commands; each failure degrades to an unreadable field."""
    ioreg = run(["/usr/sbin/ioreg", "-c", "IOHIDSystem", "-d", "4"])
    pmset = run(["/usr/bin/pmset", "-g", "ps"])
    sysctl = run(["/usr/sbin/sysctl", "-n", "kern.memorystatus_vm_pressure_level"])
    return HostState(
        parse_hid_idle_seconds(ioreg) if ioreg is not None else None,
        parse_on_ac(pmset) if pmset is not None else None,
        parse_pressure(sysctl) if sysctl is not None else "unknown",
    )
