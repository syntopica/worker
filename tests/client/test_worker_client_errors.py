import threading
from http.server import BaseHTTPRequestHandler, HTTPServer

import pytest

from worker.client.api_failure import ApiFailure
from worker.client.worker_client import WorkerClient


def serve(payload):
    class Stub(BaseHTTPRequestHandler):
        def do_GET(self):
            self.send_response(502)
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

        def log_message(self, format, *args):  # noqa: A002
            """Silent."""

    server = HTTPServer(("127.0.0.1", 0), Stub)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    return server


@pytest.mark.parametrize("payload", [b"<html>bad gateway</html>", b"[1, 2]", b""])
def test_non_object_error_body_is_unknown(payload):
    server = serve(payload)
    try:
        client = WorkerClient(f"http://127.0.0.1:{server.server_address[1]}", "t")
        with pytest.raises(ApiFailure) as error:
            client.get("j1")
    finally:
        server.shutdown()
    assert (error.value.status, error.value.code) == (502, "unknown")
