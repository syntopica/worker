"""When a CodexBar usage report says a provider's quota comes back, if it is spent."""

import datetime
from typing import Any

_SPENT_PERCENT = 95


def spent_until(usage: dict[str, Any], windows: tuple[str, ...], now: float) -> float | None:
    """The latest reset among spent windows, or None when none is known to be spent.

    With ``windows`` only the labelled extra windows whose id or title
    contains one of them are read (one provider can meter model families
    separately); otherwise the unlabelled primary, secondary and tertiary
    windows. A window whose usage is unknown is no evidence either way.
    """
    if windows:
        read = [
            extra.get("window") or {}
            for extra in usage.get("extraRateWindows") or []
            if any(
                w.lower() in f"{extra.get('id', '')} {extra.get('title', '')}".lower()
                for w in windows
            )
        ]
    else:
        read = [usage.get(k) or {} for k in ("primary", "secondary", "tertiary")]
    ends: list[float] = []
    for window in read:
        if window.get("usageKnown") is False or (window.get("usedPercent") or 0) < _SPENT_PERCENT:
            continue
        resets = window.get("resetsAt")
        if not isinstance(resets, str):
            continue
        end = datetime.datetime.fromisoformat(resets.replace("Z", "+00:00")).timestamp()
        if end > now:
            ends.append(end)
    return max(ends) if ends else None
