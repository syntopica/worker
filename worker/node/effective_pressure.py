"""Combine the kernel's pressure level with free memory (spec amendment 2026-09-30)."""


def effective_pressure(level: str, free_pct: float | None, min_free_pct: float) -> str:
    """``warn`` counts only below ``min_free_pct``; the kernel raises it with a third free.

    ``critical`` always counts, and an unreadable free percentage is ``unknown``.
    """
    if level == "critical":
        return level
    if level not in {"normal", "warn"} or free_pct is None:
        return "unknown"
    if level == "warn" and free_pct < min_free_pct:
        return "warn"
    return "normal"
