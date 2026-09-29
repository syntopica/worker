import socket
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import pytest

from worker.node.coordinator_link import CoordinatorLink


@pytest.fixture
def link():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    return CoordinatorLink(f"http://127.0.0.1:{port}", "token-a", "node-a")


def test_unreachable_coordinator_never_raises(link):
    assert link.lease(None, False, 44.0, 900.0) is None
    assert link.heartbeat("a", 1, False) is True
    assert link.complete("a", 1, {"outcome": "failed"}) is None
    assert link.report({"reason": None}) is None


class Answer(BaseHTTPRequestHandler):
    status = 200

    def do_POST(self):
        self.rfile.read(int(self.headers.get("Content-Length") or 0))
        self.send_response(self.status)
        self.send_header("Content-Length", "0")
        self.end_headers()

    def log_message(self, *args):
        pass


def serve(status):
    handler = type("H", (Answer,), {"status": status})
    server = ThreadingHTTPServer(("127.0.0.1", 0), handler)
    threading.Thread(
        target=server.serve_forever, kwargs={"poll_interval": 0.02}, daemon=True
    ).start()
    return server, CoordinatorLink(f"http://127.0.0.1:{server.server_address[1]}", "t", "node-a")


@pytest.mark.parametrize(("status", "alive"), [(409, False), (200, True), (500, True)])
def test_heartbeat_is_false_only_when_fenced(status, alive):
    server, link = serve(status)
    try:
        assert link.heartbeat("a", 1, False) is alive
    finally:
        server.shutdown()
        server.server_close()


def test_lease_204_is_none():
    server, link = serve(204)
    try:
        assert link.lease(None, False, 44.0, 900.0) is None
    finally:
        server.shutdown()
        server.server_close()
