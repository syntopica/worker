import pytest

from tests.api.test_api_roundtrip import api, call  # noqa: F401
from tests.ladder.test_on_demand_tasks import setup_task
from worker.node.coordinator_link import CoordinatorLink


@pytest.fixture
def config(tmp_path):
    return setup_task(tmp_path)[0]


def test_explicit_task_override_completes_through_the_authenticated_api(api, tmp_path):  # noqa: F811
    base, tokens = api
    _, payload = setup_task(tmp_path)
    status, submitted = call(base, tokens["pa"], "POST", "/v1/jobs", payload)
    assert status == 201
    request = {
        "node": "node-a",
        "kind": "task",
        "profile": "max",
        "user_active": True,
        "current_idle_s": 0,
    }
    status, leased = call(base, tokens["node-a"], "POST", "/v1/leases", request)
    assert status == 200
    assert leased["job_id"] == submitted["id"]
    assert leased["model"] == "max"
    report = {
        "generation": leased["generation"],
        "outcome": "succeeded",
        "output": {"text": '{"ok":true}', "json": {"ok": True}},
        "usage": {},
        "executor": {"node": "node-a", "provider": "max-lane", "model": "claude-x"},
        "wall_s": 1,
    }
    path = f"/v1/attempts/{leased['attempt_id']}/complete"
    assert call(base, tokens["node-a"], "POST", path, report)[0] == 200
    assert call(base, tokens["node-a"], "POST", path, report)[0] == 409


@pytest.mark.parametrize(
    "kind,profile", [("task", "source"), ("task", "missing"), ("inference", "max")]
)
def test_only_explicit_on_demand_task_profiles_are_accepted(api, kind, profile):  # noqa: F811
    base, tokens = api
    request = {
        "node": "node-a",
        "kind": kind,
        "profile": profile,
        "user_active": False,
        "current_idle_s": 0,
    }
    status, error = call(base, tokens["node-a"], "POST", "/v1/leases", request)
    assert (status, error["error"]) == (400, "unknown_profile")


@pytest.mark.parametrize("cutoff", [True, -1, "now", [], float("nan"), float("inf"), 10**400])
def test_a_snapshot_rejects_invalid_creation_times(api, cutoff):  # noqa: F811
    base, tokens = api
    request = {
        "node": "node-a",
        "kind": "task",
        "profile": "max",
        "user_active": False,
        "current_idle_s": 0,
        "created_before": cutoff,
    }
    assert call(base, tokens["node-a"], "POST", "/v1/leases", request)[0] == 400


def test_the_node_link_transmits_the_snapshot_boundary(api, tmp_path):  # noqa: F811
    base, tokens = api
    _, payload = setup_task(tmp_path)
    assert call(base, tokens["pa"], "POST", "/v1/jobs", payload)[0] == 201
    link = CoordinatorLink(base, tokens["node-a"], "node-a", created_before=0.0)
    assert link.lease_task(False, 0, "max") is None
    unbounded = CoordinatorLink(base, tokens["node-a"], "node-a")
    assert unbounded.lease_task(False, 0, "max") is not None
