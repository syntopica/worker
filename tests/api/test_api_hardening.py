import http.client
import stat

import pytest

from tests.api.test_api_roundtrip import api, call  # noqa: F401
from worker.api import make_handler as make_handler_module
from worker.auth.add_principal import add_principal

ROUTES = [
    ("POST", "/v1/jobs"),
    ("GET", "/v1/jobs/x"),
    ("POST", "/v1/jobs/x/cancel"),
    ("POST", "/v1/jobs/x/ack"),
    ("GET", "/v1/results"),
    ("POST", "/v1/leases"),
    ("POST", "/v1/attempts/x/heartbeat"),
    ("POST", "/v1/attempts/x/complete"),
    ("POST", "/v1/nodes/node-a/report"),
    ("GET", "/v1/status"),
]


@pytest.mark.parametrize(("method", "path"), ROUTES)
def test_oversized_content_length_is_413_without_reading_the_body(api, config, method, path):  # noqa: F811
    base, t = api
    conn = http.client.HTTPConnection(base.removeprefix("http://"), timeout=10)
    conn.putrequest(method, path)
    conn.putheader("Authorization", f"Bearer {t['pa']}")
    conn.putheader("Content-Length", str(config.max_payload_bytes + 1))
    conn.endheaders()  # the announced body is never sent
    response = conn.getresponse()
    assert (response.status, response.read()) == (413, b'{"error": "payload_too_large"}')
    conn.close()


def test_payload_at_the_limit_is_not_refused_for_size(api, config):  # noqa: F811
    base, t = api
    status, _ = call(base, t["pa"], "POST", "/v1/jobs", b"x" * config.max_payload_bytes)
    assert status != 413


@pytest.mark.parametrize(
    ("path", "payload"),
    [
        ("/v1/jobs/x/ack", b"[]"),  # AttributeError
        ("/v1/jobs/x/ack", b"{not json"),  # ValueError
        ("/v1/jobs/x/ack", b"5"),  # AttributeError on int
    ],
)
def test_malformed_bodies_are_400(api, path, payload):  # noqa: F811
    base, t = api
    assert call(base, t["pa"], "POST", path, payload) == (400, {"error": "bad_request"})


def test_missing_and_mistyped_fields_are_400(api):  # noqa: F811
    base, t = api
    assert call(base, t["node-a"], "POST", "/v1/attempts/x/heartbeat", {})[0] == 400
    assert (
        call(base, t["node-a"], "POST", "/v1/attempts/x/heartbeat", {"generation": None})[0] == 400
    )


def test_unexpected_exception_is_500_and_leaks_nothing(api, monkeypatch, capsys):  # noqa: F811
    base, t = api

    def boom(ctx):
        raise RuntimeError("SECRET-GENERATED-TEXT")

    monkeypatch.setattr(make_handler_module, "route_request", boom)
    status, payload = call(base, t["pa"], "GET", "/v1/jobs/x")
    captured = capsys.readouterr()
    assert (status, payload) == (500, {"error": "internal"})
    assert "SECRET-GENERATED-TEXT" not in captured.out + captured.err
    assert "Traceback" not in captured.out + captured.err


def test_unknown_route_is_404(api):  # noqa: F811
    base, t = api
    assert call(base, t["pa"], "GET", "/v1/nope") == (404, {"error": "no_route"})


def test_principal_files_are_owner_only(tmp_path):
    state = tmp_path / "state"
    add_principal(state, "producer", "pa")
    assert stat.S_IMODE((state / "principals.json").stat().st_mode) == 0o600
    assert stat.S_IMODE((state / "tokens" / "pa.token").stat().st_mode) == 0o600
