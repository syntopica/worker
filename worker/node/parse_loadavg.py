"""Read the one-minute load average from ``/proc/loadavg``."""


def parse_loadavg(text: str) -> float | None:
    """The first field, or None when the text is not a load average."""
    fields = text.split()
    try:
        return float(fields[0])
    except (IndexError, ValueError):
        return None
