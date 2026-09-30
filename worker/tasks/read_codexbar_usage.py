"""Ask CodexBar for one provider's usage report."""

import json
import subprocess
from typing import Any

_TIMEOUT_S = 90


def read_codexbar_usage(provider: str) -> dict[str, Any] | None:
    """The ``usage`` object, or None when CodexBar is missing, slow or unreadable."""
    try:
        done = subprocess.run(
            ["codexbar", "usage", "--provider", provider, "--format", "json"],
            capture_output=True,
            text=True,
            timeout=_TIMEOUT_S,
            check=False,
        )
        usage = json.loads(done.stdout)[0]["usage"]
    except (OSError, subprocess.TimeoutExpired, ValueError, KeyError, IndexError, TypeError):
        return None
    return usage if isinstance(usage, dict) else None
