"""Map ``sysctl -n kern.memorystatus_vm_pressure_level`` to a named level."""

_LEVELS = {"1": "normal", "2": "warn", "4": "critical"}


def parse_pressure(text: str) -> str:
    """Anything unexpected is ``unknown``, which blocks work."""
    return _LEVELS.get(text.strip(), "unknown")
