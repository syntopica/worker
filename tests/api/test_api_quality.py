from tests.api.test_api_roundtrip import api, call  # noqa: F401
from tests.conftest import body
from worker.client.worker_client import WorkerClient

LEASE = {
    "node": "node-a",
    "resident_model": None,
    "user_active": False,
    "free_gb": 44,
    "current_idle_s": 900,
}


def run_one(base, t):
    _, sub = call(base, t["pa"], "POST", "/v1/jobs", body(tier="strong"))
    _, lease = call(base, t["node-a"], "POST", "/v1/leases", LEASE)
    report = {
        "generation": lease["generation"],
        "outcome": "succeeded",
        "output": {"text": "ok", "json": None},
        "usage": {},
        "executor": {"node": "node-a", "provider": "ollama", "model": "model-a"},
        "error_code": None,
        "wall_s": 3.0,
    }
    call(base, t["node-a"], "POST", f"/v1/attempts/{lease['attempt_id']}/complete", report)
    _, results = call(base, t["pa"], "GET", "/v1/results?queue=pa.bulk&after=0&wait=1")
    return sub["id"], results["results"][0]["result_id"]


def test_a_rated_ack_shows_in_the_quality_report(api):  # noqa: F811
    base, t = api
    job_id, result_id = run_one(base, t)
    ack = {"result_id": result_id, "rating": "discarded"}
    assert call(base, t["pa"], "POST", f"/v1/jobs/{job_id}/ack", ack)[0] == 200
    assert call(base, t["pa"], "GET", "/v1/quality")[0] == 403
    status, report = call(base, t["admin"], "GET", "/v1/quality?days=1")
    assert status == 200
    [attempt] = report["attempts"]
    assert (attempt["tier"], attempt["model"], attempt["succeeded"]) == ("strong", "model-a", 1)
    [rating] = report["ratings"]
    assert (rating["discarded"], rating["rated"]) == (1, 1)


def test_an_unknown_rating_is_refused(api):  # noqa: F811
    base, t = api
    job_id, result_id = run_one(base, t)
    status, error = call(
        base, t["pa"], "POST", f"/v1/jobs/{job_id}/ack", {"result_id": result_id, "rating": "meh"}
    )
    assert (status, error["error"]) == (400, "unknown_rating")


def test_the_client_sends_a_rating_with_its_ack(api):  # noqa: F811
    base, t = api
    job_id, result_id = run_one(base, t)
    WorkerClient(base, t["pa"]).ack(job_id, result_id, rating="good")
    _, report = call(base, t["admin"], "GET", "/v1/quality")
    assert report["ratings"][0]["good"] == 1
