"""Read and drop a refused request body so the client can read the refusal."""

import io
import socket

_CHUNK = 65536
_MAX_DISCARD = 64 * 1024 * 1024
_IDLE_S = 0.5


def discard_body(rfile: io.BufferedIOBase, sock: socket.socket, length: int) -> None:
    """Drain up to ``length`` bytes (at most 64 MiB) while they keep arriving.

    Closing with the body unread makes the kernel reset the connection, and a
    client still writing sees ``Connection reset by peer`` instead of the 413
    (Atrium hit this with oversized prompts on 2026-09-30). A client that
    sends nothing for half a second is not waited for.
    """
    left = min(length, _MAX_DISCARD)
    sock.settimeout(_IDLE_S)
    try:
        while left > 0:
            chunk = rfile.read1(min(_CHUNK, left))
            if not chunk:
                return
            left -= len(chunk)
    except OSError:
        return
