import threading
from http.server import BaseHTTPRequestHandler, HTTPServer

import pytest

from worker.client.worker_client import WorkerClient

RAW_ID = "../v1/status"
QUOTED = "..%2Fv1%2Fstatus"


def recording_server(paths):
    class Stub(BaseHTTPRequestHandler):
        def _answer(self):
            paths.append(self.path)
            payload = b'{"state": "cancelled"}'
            self.send_response(200)
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

        do_GET = _answer  # noqa: N815
        do_POST = _answer  # noqa: N815

        def log_message(self, format, *args):  # noqa: A002
            """Silent."""

    server = HTTPServer(("127.0.0.1", 0), Stub)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    return server


@pytest.mark.parametrize(
    ("call", "suffix"),
    [
        (lambda c: c.ack(RAW_ID, "r1"), "/ack"),
        (lambda c: c.cancel(RAW_ID), "/cancel"),
    ],
)
def test_job_id_is_quoted_in_mutating_paths(call, suffix):
    paths = []
    server = recording_server(paths)
    try:
        call(WorkerClient(f"http://127.0.0.1:{server.server_address[1]}", "t"))
    finally:
        server.shutdown()
    assert paths == [f"/v1/jobs/{QUOTED}{suffix}"]


def test_job_id_is_quoted_in_get_path():
    paths = []
    server = recording_server(paths)
    try:
        client = WorkerClient(f"http://127.0.0.1:{server.server_address[1]}", "t")
        client.get(RAW_ID)
    finally:
        server.shutdown()
    assert paths == [f"/v1/jobs/{QUOTED}"]
