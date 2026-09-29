import socket

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
