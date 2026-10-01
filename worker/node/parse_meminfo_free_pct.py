"""Read the available share of memory from ``/proc/meminfo``."""


def parse_meminfo_free_pct(text: str) -> float | None:
    """``MemAvailable`` over ``MemTotal`` in percent, or None when either is missing."""
    values: dict[str, float] = {}
    for line in text.splitlines():
        name, _, rest = line.partition(":")
        parts = rest.split()
        if name in {"MemTotal", "MemAvailable"} and parts and parts[0].isdigit():
            values[name] = float(parts[0])
    total, available = values.get("MemTotal"), values.get("MemAvailable")
    if not total or available is None:
        return None
    return 100.0 * available / total
