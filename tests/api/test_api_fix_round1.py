import json
import socket
import stat

import pytest

from tests.api.test_api_roundtrip import api, call  # noqa: F401
from worker.api import handle_results as results_module
from worker.api.handle_results import handle_results
from worker.api.make_handler import make_handler
from worker.api.request_context import RequestContext
from worker.auth.add_principal import add_principal
from worker.auth.principal import Principal


def test_wait_nan_is_400(api):  # noqa: F811
    base, t = api
    assert call(base, t["pa"], "GET", "/v1/results?queue=pa.bulk&wait=nan") == (
        400,
        {"error": "bad_request"},
    )
    assert call(base, t["pa"], "GET", "/v1/results?queue=pa.bulk&wait=inf")[0] == 400


def test_wait_is_clamped_to_30_seconds(config, conn, monkeypatch):
    clock = {"now": 0.0}
    monkeypatch.setattr(results_module.time, "monotonic", lambda: clock["now"])
    monkeypatch.setattr(results_module.time, "sleep", lambda s: clock.update(now=clock["now"] + s))
    for wait, bound in (("999", 30.5), ("-5", 0.5)):
        clock["now"] = 0.0
        ctx = RequestContext(
            Principal("producer", "pa"),
            "GET",
            ["v1", "results"],
            {"queue": "pa.bulk", "wait": wait},
            b"",
            config,
            conn,
            0.0,
        )
        assert handle_results(ctx) == (200, {"results": []})
        assert clock["now"] <= bound


def test_node_report_round_trips_into_status(api):  # noqa: F811
    base, t = api
    assert (
        call(
            base,
            t["node-a"],
            "POST",
            "/v1/nodes/node-a/report",
            {"reason": "idle", "model": "model-a"},
        )[0]
        == 200
    )
    status, payload = call(base, t["admin"], "GET", "/v1/status")
    assert status == 200
    assert payload["nodes"]["node-a"]["reason"] == "idle"


@pytest.mark.parametrize("payload", [b"[]", b"5", b"", b"{oops", b"null"])
def test_node_report_must_be_a_json_object(api, payload):  # noqa: F811
    base, t = api
    assert call(base, t["node-a"], "POST", "/v1/nodes/node-a/report", payload) == (
        400,
        {"error": "bad_request"},
    )
    assert call(base, t["admin"], "GET", "/v1/status")[0] == 200


def test_node_cannot_report_for_another_node(api):  # noqa: F811
    base, t = api
    assert call(base, t["node-a"], "POST", "/v1/nodes/node-g/report", {"reason": "x"})[0] == 403


def test_producer_lease_is_403_before_the_body_is_parsed(api):  # noqa: F811
    base, t = api
    assert call(base, t["pa"], "POST", "/v1/leases", b"{not json") == (403, {"error": "forbidden"})


def test_handler_has_a_socket_timeout(config, tmp_path):
    assert make_handler(config, tmp_path / "state").timeout == 30


def test_stalled_client_does_not_block_other_requests(api):  # noqa: F811
    base, t = api
    host, port = base.removeprefix("http://").split(":")
    stalled = socket.create_connection((host, int(port)))
    stalled.sendall(b"POST /v1/jobs HTTP/1.1\r\nContent-Length: 10\r\n\r\n")
    assert call(base, t["admin"], "GET", "/v1/status")[0] == 200
    stalled.close()


def test_principal_files_stay_private_after_rewrite(tmp_path):
    state = tmp_path / "state"
    add_principal(state, "producer", "pa")
    (state / "tokens").chmod(0o755)
    add_principal(state, "node", "node-a")
    assert stat.S_IMODE((state / "tokens").stat().st_mode) == 0o700
    assert stat.S_IMODE((state / "principals.json").stat().st_mode) == 0o600
    assert not list(state.glob("*.tmp"))
    assert set(json.loads((state / "principals.json").read_text())) == {
        "pa",
        "node-a",
    }
