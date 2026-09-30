"""Write a prompt to a runner's stdin and close it, ignoring a runner that left."""

import contextlib
from typing import IO


def feed_stdin(pipe: IO[bytes], text: str) -> None:
    """Run on a thread: a pipe write blocks until the runner reads."""
    with contextlib.suppress(OSError):
        pipe.write(text.encode())
    with contextlib.suppress(OSError):
        pipe.close()
