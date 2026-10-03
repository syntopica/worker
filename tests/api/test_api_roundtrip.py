import json
import threading
import urllib.error
import urllib.request

import pytest

from tests.conftest import body
from worker.api.build_server import build_server
from worker.auth.add_principal import add_principal
from worker.store.migrate_state import migrate_state


@pytest.fixture
def api(config, tmp_path):
    state = tmp_path / "state"
    tokens = {
        name: add_principal(state, kind, name)
        for kind, name in (("producer", "pa"), ("node", "node-a"), ("admin", "admin"))
    }
    migrate_state(state)
    server = build_server(config, state)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    base = f"http://127.0.0.1:{server.server_address[1]}"
    yield base, tokens
    server.shutdown()


def call(base, token, method, path, payload=None):
    data = (
        payload
        if isinstance(payload, bytes)
        else (json.dumps(payload).encode() if payload is not None else None)
    )
    request = urllib.request.Request(
        base + path,
        data=data,
        method=method,
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(request, timeout=40) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else None
    except urllib.error.HTTPError as error:
        return error.code, json.loads(error.read())


def test_submit_lease_complete_collect_ack(api):
    base, t = api
    status, sub = call(base, t["pa"], "POST", "/v1/jobs", body())
    assert status == 201
    status, lease = call(
        base,
        t["node-a"],
        "POST",
        "/v1/leases",
        {
            "node": "node-a",
            "resident_model": None,
            "user_active": False,
            "free_gb": 44,
            "current_idle_s": 900,
        },
    )
    assert (status, lease["job_id"]) == (200, sub["id"])
    report = {
        "generation": lease["generation"],
        "outcome": "succeeded",
        "output": {"text": "ok", "json": None},
        "usage": {},
        "executor": {"node": "node-a"},
        "error_code": None,
        "wall_s": 1.0,
    }
    assert (
        call(base, t["node-a"], "POST", f"/v1/attempts/{lease['attempt_id']}/complete", report)[0]
        == 200
    )
    status, results = call(base, t["pa"], "GET", "/v1/results?queue=pa.bulk&after=0&wait=1")
    assert results["results"][0]["output"]["text"] == "ok"
    assert (
        call(
            base,
            t["pa"],
            "POST",
            f"/v1/jobs/{sub['id']}/ack",
            {"result_id": results["results"][0]["result_id"]},
        )[0]
        == 200
    )


def test_wrong_principal_and_stale_attempt_are_refused(api):
    base, t = api
    assert call(base, "nope", "POST", "/v1/jobs", body())[0] == 401
    assert call(base, t["pa"], "POST", "/v1/leases", {})[0] == 403
    status, err = call(
        base,
        t["node-a"],
        "POST",
        "/v1/attempts/none/heartbeat",
        {"generation": 1, "draining": False},
    )
    assert (status, err["error"]) == (409, "stale_attempt")


def test_status_needs_admin(api):
    base, t = api
    assert call(base, t["pa"], "GET", "/v1/status")[0] == 403
    assert call(base, t["admin"], "GET", "/v1/status")[0] == 200


def test_activity_needs_admin_and_a_numeric_span(api):
    base, t = api
    assert call(base, t["pa"], "GET", "/v1/activity")[0] == 403
    status, data = call(base, t["admin"], "GET", "/v1/activity?hours=168")
    assert status == 200
    assert data["bucket_s"] == 21600.0
    assert data["rows"] == []
    assert call(base, t["admin"], "GET", "/v1/activity?hours=lots")[0] == 400
