"""A signal handler that raises ``SystemExit`` in the main thread."""

import sys
from types import FrameType


def exit_on_signal(_signum: int, _frame: FrameType | None) -> None:
    """Exit cleanly, so ``except SystemExit`` blocks can hand work back."""
    sys.exit(0)
