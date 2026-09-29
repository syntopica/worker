import json

from tests.conftest import body
from worker.cli.read_token import read_token
from worker.node.coordinator_link import CoordinatorLink


def test_submit_jobs_cancel_and_status_round_trip(served, run):
    for kind, name in (("producer", "pa"), ("admin", "admin")):
        assert run("token", "add", "--kind", kind, "--name", name)[0] == 0
    job_file = served / "job.json"
    job_file.write_bytes(body())

    code, out, _ = run("submit", "--producer", "pa", str(job_file))
    submitted = json.loads(out)
    assert (code, submitted["created"]) == (0, True)

    code, out, _ = run("jobs", "--producer", "pa", "--queue", "pa.bulk")
    assert (code, json.loads(out)) == (0, [])

    code, out, _ = run("cancel", "--producer", "pa", submitted["id"])
    assert (code, json.loads(out)) == (0, {"state": "cancelled"})

    code, out, _ = run("status", "--json")
    status = json.loads(out)
    assert code == 0
    assert status["queues"]["pa.bulk"]["states"] == {"cancelled": 1}


def test_jobs_lists_a_real_unacked_result(served, run):
    for kind, name in (("producer", "pa"), ("node", "node-a")):
        run("token", "add", "--kind", kind, "--name", name)
    (served / "job.json").write_bytes(body())
    job_id = json.loads(run("submit", "--producer", "pa", str(served / "job.json"))[1])["id"]
    state = served / "worker" / "state"
    listen = json.loads((served / "worker" / "config.json").read_text())["listen"]
    link = CoordinatorLink(f"http://{listen}", read_token(state, "node-a"), "node-a")
    lease = link.lease(None, False, 44.0, 900.0)
    assert lease is not None
    link.complete(
        lease["attempt_id"],
        lease["generation"],
        {
            "outcome": "succeeded",
            "output": {"text": "ok", "json": None},
            "usage": {},
            "executor": {"node": "node-a"},
            "error_code": None,
            "wall_s": 1.0,
        },
    )
    code, out, _ = run("jobs", "--producer", "pa", "--queue", "pa.bulk", "--json")
    results = json.loads(out)
    assert code == 0
    assert [(r["job_id"], r["output"]["text"]) for r in results] == [(job_id, "ok")]


def test_a_refusal_is_one_stderr_line_and_exit_1(served, run):
    run("token", "add", "--kind", "producer", "--name", "pa")
    assert run("token", "add", "--kind", "producer", "--name", "admin")[0] == 0
    code, out, err = run("status")
    assert (code, out, err) == (1, "", "worker: 403 forbidden\n")


def test_cancelling_an_unknown_job_reports_the_code(served, run):
    run("token", "add", "--kind", "producer", "--name", "pa")
    code, _, err = run("cancel", "--producer", "pa", "nope")
    assert code == 1
    assert err.startswith("worker: 404 ")
