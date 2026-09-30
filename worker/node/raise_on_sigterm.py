"""Turn launchd's SIGTERM into ``SystemExit`` so a running attempt can hand its job back."""

import signal

from worker.node.exit_on_signal import exit_on_signal


def raise_on_sigterm() -> None:
    """Install the handler; call once from the main thread of a loop command."""
    signal.signal(signal.SIGTERM, exit_on_signal)
