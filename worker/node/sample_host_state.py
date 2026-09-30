"""Read idle time, power source and memory pressure."""

from collections.abc import Callable

from worker.config.default_min_free_pct import DEFAULT_MIN_FREE_PCT
from worker.node.effective_pressure import effective_pressure
from worker.node.host_state import HostState
from worker.node.parse_free_pct import parse_free_pct
from worker.node.parse_hid_idle_seconds import parse_hid_idle_seconds
from worker.node.parse_on_ac import parse_on_ac
from worker.node.parse_pressure import parse_pressure
from worker.node.run_command import run_command


def sample_host_state(
    run: Callable[[list[str]], str | None] = run_command,
    min_free_pct: float = DEFAULT_MIN_FREE_PCT,
) -> HostState:
    """Four short commands; each failure degrades to an unreadable field."""
    ioreg = run(["/usr/sbin/ioreg", "-c", "IOHIDSystem", "-d", "4"])
    pmset = run(["/usr/bin/pmset", "-g", "ps"])
    level = run(["/usr/sbin/sysctl", "-n", "kern.memorystatus_vm_pressure_level"])
    free = run(["/usr/sbin/sysctl", "-n", "kern.memorystatus_level"])
    return HostState(
        parse_hid_idle_seconds(ioreg) if ioreg is not None else None,
        parse_on_ac(pmset) if pmset is not None else None,
        effective_pressure(
            parse_pressure(level) if level is not None else "unknown",
            parse_free_pct(free) if free is not None else None,
            min_free_pct,
        ),
    )
