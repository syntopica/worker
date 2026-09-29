import pytest

from worker.cli.base_url import base_url


class Cfg:
    def __init__(self, host):
        self.listen_host = host
        self.listen_port = 8765


@pytest.mark.parametrize(
    ("host", "expected"),
    [
        ("127.0.0.1", "http://127.0.0.1:8765"),
        ("0.0.0.0", "http://127.0.0.1:8765"),  # noqa: S104
        ("::", "http://[::1]:8765"),
        ("::1", "http://[::1]:8765"),
        ("fe80::1", "http://[fe80::1]:8765"),
        ("localhost", "http://localhost:8765"),
    ],
)
def test_base_url(host, expected):
    assert base_url(Cfg(host)) == expected
