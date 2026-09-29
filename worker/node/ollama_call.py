"""One /api/chat request on a thread, cancellable by closing its socket."""

import http.client
import json
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
            self._conn.request("POST", "/api/chat", data, {"Content-Type": "application/json"})
            response = self._conn.getresponse()
            payload = json.loads(response.read())
            if response.status != http.HTTPStatus.OK or "error" in payload:
                self._error = f"http_{response.status}"
            else:
                self._answer = payload
        except (OSError, http.client.HTTPException, ValueError):
            self._error = "cancelled" if self._cancelled else "transport_error"

    def done(self) -> bool:
        """True once the thread has finished, for any reason."""
        return self._thread is not None and not self._thread.is_alive()

    def cancel(self) -> None:
        """Close the socket so the blocked read fails."""
        self._cancelled = True
        if self._conn.sock is not None:
            self._conn.sock.close()
        self._conn.close()

    def outcome(self) -> tuple[dict[str, Any] | None, str | None]:
        """(answer, None) on success, (None, code) otherwise."""
        if self._cancelled:
            return None, "cancelled"
        return self._answer, self._error
