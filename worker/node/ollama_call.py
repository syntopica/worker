"""One /api/chat request on a thread, cancellable by closing its socket."""

import contextlib
import http.client
import json
import socket
import threading
import urllib.parse
from typing import Any


class OllamaCall:
    """Closing the connection is only a hint to Ollama; drain_backend confirms quiet."""

    def __init__(self, conn: http.client.HTTPConnection) -> None:
        self._conn = conn
        self._answer: dict[str, Any] | None = None
        self._error: str | None = None
        self._cancelled = False
        self._thread: threading.Thread | None = None

    @classmethod
    def start(cls, url: str, body: dict[str, Any], timeout: float) -> "OllamaCall":
        """Send the request in the background and return immediately."""
        parts = urllib.parse.urlsplit(url)
        call = cls(
            http.client.HTTPConnection(
                parts.hostname or "127.0.0.1", parts.port or 11434, timeout=timeout
            )
        )
        call._thread = threading.Thread(
            target=call._run, args=(json.dumps(body).encode(),), daemon=True
        )
        call._thread.start()
        return call

    def _run(self, data: bytes) -> None:
        try:
            if self._was_cancelled():
                return
            self._conn.connect()
            if self._was_cancelled():
                return
            self._conn.request("POST", "/api/chat", data, {"Content-Type": "application/json"})
            response = self._conn.getresponse()
            body = response.read()
        except (OSError, http.client.HTTPException, AttributeError):
            self._error = "cancelled" if self._cancelled else "transport_error"
            return
        if response.status != http.HTTPStatus.OK:
            self._error = f"http_{response.status}"
            return
        try:
            payload = json.loads(body)
        except ValueError:
            payload = None
        if isinstance(payload, dict) and "error" not in payload:
            self._answer = payload
        else:
            self._error = "bad_response"

    def _was_cancelled(self) -> bool:
        return self._cancelled

    def done(self) -> bool:
        """True once the thread has finished, for any reason."""
        return self._thread is not None and not self._thread.is_alive()

    def cancel(self) -> None:
        """Close the socket so the blocked read fails."""
        self._cancelled = True
        sock = self._conn.sock
        if sock is not None:
            with contextlib.suppress(OSError):
                sock.shutdown(socket.SHUT_RDWR)
        with contextlib.suppress(OSError):
            self._conn.close()

    def outcome(self) -> tuple[dict[str, Any] | None, str | None]:
        """(answer, None) on success, (None, code) otherwise."""
        if self._cancelled:
            return None, "cancelled"
        return self._answer, self._error
