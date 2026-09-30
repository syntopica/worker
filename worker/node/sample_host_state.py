"""Read idle time, power source and memory pressure."""

import sys
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
    """Four short commands; each failure degrades to an unreadable field.

    A failed reader is logged by its fixed name only, to tell which one makes
    ``host_state_unreadable`` flap.
    """
    ioreg = run(["/usr/sbin/ioreg", "-c", "IOHIDSystem", "-d", "4"])
    pmset = run(["/usr/bin/pmset", "-g", "ps"])
    level = run(["/usr/sbin/sysctl", "-n", "kern.memorystatus_vm_pressure_level"])
    free = run(["/usr/sbin/sysctl", "-n", "kern.memorystatus_level"])
    idle = parse_hid_idle_seconds(ioreg) if ioreg is not None else None
    on_ac = parse_on_ac(pmset) if pmset is not None else None
    pressure = parse_pressure(level) if level is not None else "unknown"
    checks = (
        ("ioreg", ioreg, idle),
        ("pmset", pmset, on_ac),
        ("pressure_level", level, None if pressure == "unknown" else pressure),
    )
    failed = [n if raw is None else f"{n}_parse" for n, raw, value in checks if value is None]
    if failed:
        print(f"worker: host readers failed: {','.join(failed)}", file=sys.stderr)
    return HostState(
        idle,
        on_ac,
        effective_pressure(
            pressure, parse_free_pct(free) if free is not None else None, min_free_pct
        ),
    )
