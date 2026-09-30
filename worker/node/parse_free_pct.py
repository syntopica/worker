"""Parse ``sysctl -n kern.memorystatus_level``, the free-memory percentage."""


def parse_free_pct(text: str) -> float | None:
    """None for anything that is not a percentage, which blocks work."""
    try:
        value = float(text.strip())
    except ValueError:
        return None
    return value if 0 <= value <= 100 else None  # noqa: PLR2004
