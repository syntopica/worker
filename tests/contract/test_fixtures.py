import json
import threading
import urllib.error
import urllib.request
from pathlib import Path

import pytest

from worker.api.build_server import build_server
from worker.auth.add_principal import add_principal
from worker.jobs.parse_submit_request import parse_submit_request

FIXTURES = Path(__file__).parents[2] / "contract" / "fixtures"
LEASE_REQUEST = {
    "node": "node-a",
    "resident_model": None,
    "user_active": False,
    "free_gb": 44,
    "current_idle_s": 900,
}


def fixture(name):
    return json.loads((FIXTURES / name).read_text())


def call(base, token, method, path, payload=None):
    data = json.dumps(payload).encode() if payload is not None else None
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    request = urllib.request.Request(base + path, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=40) as response:
            return response.status, json.loads(response.read())
    except urllib.error.HTTPError as error:
        return error.code, json.loads(error.read())


@pytest.fixture
def api(config, tmp_path):
    state = tmp_path / "state"
    tokens = {
        name: add_principal(state, kind, name)
        for kind, name in (("producer", "pa"), ("node", "node-a"))
    }
    server = build_server(config, state)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    yield f"http://127.0.0.1:{server.server_address[1]}", tokens
    server.shutdown()


def test_submit_fixture_is_a_valid_v1_request():
    request = parse_submit_request(fixture("submit_inference.json"))
    assert request.privacy == "mail"


def test_result_fixtures_carry_the_documented_keys():
    keys = {"seq", "result_id", "job_id", "control", "detail", "output", "executor", "usage"}
    for name in ("result_succeeded.json", "result_split_requested.json"):
        assert set(fixture(name)) == keys


def test_fixture_key_sets_match_the_real_server(api):
    base, tokens = api
    job = {**fixture("submit_inference.json"), "queue": "pa.bulk"}
    assert call(base, tokens["pa"], "POST", "/v1/jobs", job)[0] == 201
    status, lease = call(base, tokens["node-a"], "POST", "/v1/leases", LEASE_REQUEST)
    assert status == 200
    assert set(fixture("lease.json")) == set(lease)
    stale = {"generation": lease["generation"] + 1, "outcome": "succeeded"}
    path = f"/v1/attempts/{lease['attempt_id']}/complete"
    status, error = call(base, tokens["node-a"], "POST", path, stale)
    assert (status, set(error)) == (409, set(fixture("error_stale_attempt.json")))
    assert error == fixture("error_stale_attempt.json")
    report = {
        "generation": lease["generation"],
        "outcome": "succeeded",
        "output": {"text": "ok", "json": None},
        "usage": {"tokens_in": 1, "tokens_out": 1},
        "executor": {"node": "node-a", "provider": "ollama", "model": "model-a"},
        "error_code": None,
        "wall_s": 1.0,
    }
    assert call(base, tokens["node-a"], "POST", path, report)[0] == 200
    _, results = call(base, tokens["pa"], "GET", "/v1/results?queue=pa.bulk&after=0")
    real = set(results["results"][0])
    for name in ("result_succeeded.json", "result_split_requested.json"):
        assert set(fixture(name)) == real
