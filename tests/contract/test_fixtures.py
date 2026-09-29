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


def lease_one(base, tokens):
    status, lease = call(base, tokens["node-a"], "POST", "/v1/leases", LEASE_REQUEST)
    assert status == 200
    return lease, f"/v1/attempts/{lease['attempt_id']}/complete"


def report(lease, outcome, **extra):
    body = {
        "generation": lease["generation"],
        "outcome": outcome,
        "output": None,
        "usage": {},
        "executor": {"node": "node-a", "provider": "ollama", "model": "model-a"},
        "error_code": None,
        "wall_s": 1.0,
    }
    return {**body, **extra}


def submit(base, tokens, **extra):
    job = {**fixture("submit_inference.json"), "queue": "pa.bulk", **extra}
    assert call(base, tokens["pa"], "POST", "/v1/jobs", job)[0] == 201


def first_result(base, tokens):
    _, results = call(base, tokens["pa"], "GET", "/v1/results?queue=pa.bulk&after=0")
    return results["results"][0]


def test_lease_and_stale_error_match_the_real_server(api):
    base, tokens = api
    submit(base, tokens)
    lease, path = lease_one(base, tokens)
    assert set(fixture("lease.json")) == set(lease)
    assert set(fixture("lease.json")["input"]) == set(lease["input"])
    stale = report(lease, "succeeded", generation=lease["generation"] + 1)
    status, error = call(base, tokens["node-a"], "POST", path, stale)
    assert (status, error) == (409, fixture("error_stale_attempt.json"))


def test_succeeded_fixture_matches_the_real_server(api):
    base, tokens = api
    submit(base, tokens)
    lease, path = lease_one(base, tokens)
    done = report(
        lease,
        "succeeded",
        output={"text": "ok", "json": None},
        usage={"tokens_in": 1, "tokens_out": 1},
    )
    assert call(base, tokens["node-a"], "POST", path, done)[0] == 200
    real, expected = first_result(base, tokens), fixture("result_succeeded.json")
    assert set(real) == set(expected)
    for key in ("output", "executor", "usage"):
        assert set(real[key]) == set(expected[key])


def test_split_requested_fixture_matches_the_real_server(api):
    base, tokens = api
    submit(base, tokens)
    for _ in range(3):
        lease, path = lease_one(base, tokens)
        call(base, tokens["node-a"], "POST", path, report(lease, "preempted"))
    real, expected = first_result(base, tokens), fixture("result_split_requested.json")
    assert set(real) == set(expected)
    assert real["control"] == expected["control"]
    assert (real["output"], real["executor"], real["usage"]) == (None, None, None)
    assert set(real["detail"]) == set(expected["detail"])


def test_failed_fixture_matches_the_real_server(api):
    base, tokens = api
    submit(base, tokens, max_attempts=1)
    lease, path = lease_one(base, tokens)
    failed = report(lease, "failed", error_code="transport_error")
    assert call(base, tokens["node-a"], "POST", path, failed)[0] == 200
    real, expected = first_result(base, tokens), fixture("result_failed.json")
    assert set(real) == set(expected)
    assert real["control"] == expected["control"]
    assert (real["output"], real["executor"], real["usage"]) == (None, None, None)
    assert set(real["detail"]) == set(expected["detail"])
