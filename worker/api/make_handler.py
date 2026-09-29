"""Build the BaseHTTPRequestHandler class bound to one configuration."""

import json
import time
import urllib.parse
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler
from pathlib import Path

from worker.api.request_context import RequestContext
from worker.api.route_request import route_request
from worker.auth.authenticate import authenticate
from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.store.open_store import open_store


def make_handler(config: WorkerConfig, state_dir: Path) -> type[BaseHTTPRequestHandler]:
    """Each request opens its own connection; errors carry only their code."""

    class Handler(BaseHTTPRequestHandler):
        timeout = 30  # a client that stalls mid-request cannot hold a thread forever

        def _dispatch(self) -> tuple[int, dict[str, object]]:
            length = int(self.headers.get("Content-Length") or 0)
            if length > config.max_payload_bytes:
                self.close_connection = True
                raise ApiError(413, "payload_too_large")
            url = urllib.parse.urlsplit(self.path)
            raw = self.rfile.read(max(length, 0)) if length > 0 else b""
            conn = open_store(state_dir)
            try:
                principal = authenticate(state_dir, self.headers.get("Authorization"))
                query = dict(urllib.parse.parse_qsl(url.query))
                parts = [p for p in url.path.split("/") if p]
                ctx = RequestContext(
                    principal, self.command, parts, query, raw, config, conn, time.time()
                )
                return route_request(ctx)
            finally:
                conn.close()

        def _serve(self) -> None:
            try:
                status, payload = self._dispatch()
            except ApiError as error:
                status, payload = error.status, {"error": error.code}
            except (KeyError, ValueError, TypeError, AttributeError):
                status, payload = 400, {"error": "bad_request"}
            except Exception:
                status, payload = 500, {"error": "internal"}
            data = b"" if status == HTTPStatus.NO_CONTENT else json.dumps(payload).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        do_GET = _serve  # noqa: N815
        do_POST = _serve  # noqa: N815

        def log_message(self, format: str, *args: object) -> None:  # noqa: A002
            """Metadata only: method, path shape and status, never bodies."""

    return Handler
