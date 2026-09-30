import json
import threading

from tests.conftest import body
from worker.cli.read_token import read_token
from worker.node.coordinator_link import CoordinatorLink


def _complete_next(served, output=None, outcome="succeeded", code=None):
    state = served / "worker" / "state"
    listen = json.loads((served / "worker" / "config.json").read_text())["listen"]
    link = CoordinatorLink(f"http://{listen}", read_token(state, "node-a"), "node-a")
    lease = None
    while lease is None:
        lease = link.lease(None, False, 44.0, 900.0)
    link.complete(
        lease["attempt_id"],
        lease["generation"],
        {
            "outcome": outcome,
            "output": output,
            "usage": {},
            "executor": {"node": "node-a"},
            "error_code": code,
            "wall_s": 1.0,
        },
    )


def _tokens(run):
    for kind, name in (("producer", "pa"), ("node", "node-a")):
        run("token", "add", "--kind", kind, "--name", name)


def test_run_prints_the_output_and_acknowledges_it(served, run):
    _tokens(run)
    (served / "job.json").write_bytes(body())
    node = threading.Thread(
        target=_complete_next, args=(served, {"text": '{"a": 1}', "json": {"a": 1}})
    )
    node.start()
    code, out, _ = run("run", "--producer", "pa", "--poll", "0.05", str(served / "job.json"))
    node.join()
    assert (code, json.loads(out)) == (0, {"a": 1})
    code, out, _ = run("jobs", "--producer", "pa", "--queue", "pa.bulk", "--json")
    assert json.loads(out) == []


def test_run_reports_a_pending_job_and_leaves_it_queued(served, run):
    _tokens(run)
    (served / "job.json").write_bytes(body())
    code, _, err = run("run", "--producer", "pa", "--wait", "0", str(served / "job.json"))
    assert code == 4
    assert "still queued" in err


def test_run_names_a_job_that_ended_with_nothing_to_collect(served, run):
    _tokens(run)
    (served / "job.json").write_bytes(body())
    submitted = json.loads(run("submit", "--producer", "pa", str(served / "job.json"))[1])
    run("cancel", "--producer", "pa", submitted["id"])
    code, _, err = run("run", "--producer", "pa", "--wait", "0", str(served / "job.json"))
    assert code == 1
    assert "cancelled" in err


def test_run_names_the_result_so_the_script_can_rate_it(served, run):
    _tokens(run)
    run("token", "add", "--kind", "admin", "--name", "admin")
    (served / "job.json").write_bytes(body())
    node = threading.Thread(
        target=_complete_next, args=(served, {"text": '{"a": 1}', "json": {"a": 1}})
    )
    node.start()
    code, _, err = run("run", "--producer", "pa", "--poll", "0.05", str(served / "job.json"))
    node.join()
    job_id, result_id = err.split()[2], err.split()[4]
    code, out, _ = run("rate", "--producer", "pa", job_id, result_id, "edited")
    assert (code, json.loads(out)) == (0, {"rated": "edited"})
    _, out, _ = run("quality", "--json")
    assert [r["edited"] for r in json.loads(out)["ratings"]] == [1]
