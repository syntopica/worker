"""Run a read-only system command; None on any failure (spec 7: failure = busy)."""

import subprocess


def run_command(args: list[str], timeout: float = 5.0) -> str | None:
    """Stdout on exit 0, else None."""
    try:
        done = subprocess.run(args, capture_output=True, text=True, timeout=timeout, check=False)
    except (OSError, subprocess.TimeoutExpired):
        return None
    return done.stdout if done.returncode == 0 else None
