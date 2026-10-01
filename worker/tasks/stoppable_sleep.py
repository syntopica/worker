"""A sleep that ends a slot thread once its loop is told to stop."""

import threading
from collections.abc import Callable


def stoppable_sleep(stop: threading.Event) -> Callable[[float], None]:
    """Wait up to ``seconds``; raise SystemExit as soon as ``stop`` is set.

    SystemExit takes the same path as SIGTERM in the main thread: a running
    runner's group is killed and its job handed back as a shutdown.
    """

    def sleep(seconds: float) -> None:
        if stop.wait(seconds):
            raise SystemExit(0)

    return sleep
