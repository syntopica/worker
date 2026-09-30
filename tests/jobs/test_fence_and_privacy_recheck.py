import pytest

from tests.conftest import body, fresh_store
from tests.node.test_run_attempt import Link
from tests.tasks.test_task_step import LEASE as TASK_LEASE
from tests.tasks.test_task_step import Link as TaskLink
from tests.tasks.test_task_step import make_config
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.heartbeat_attempt import heartbeat_attempt
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory
from worker.node.run_leased import run_leased
from worker.tasks.task_step import task_step


def test_only_the_lease_holder_may_heartbeat_or_complete(config, tmp_path):
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    leased = lease_job(conn, config, LeaseRequest("node-a", None, False, 40.0, 9999.0), 1.0)
    with pytest.raises(ApiError) as caught:
        heartbeat_attempt(conn, leased.attempt_id, leased.generation, False, 2.0, node="node-g")
    assert caught.value.code == "stale_attempt"
    report = CompletionReport("succeeded", {"text": "x"}, {}, {}, None, 1.0)
    with pytest.raises(ApiError):
        complete_attempt(
            conn, config, leased.attempt_id, leased.generation, report, 2.0, node="node-g"
        )
    heartbeat_attempt(conn, leased.attempt_id, leased.generation, False, 2.0, node="node-a")


def test_the_node_refuses_an_inference_lease_its_policy_forbids(config):
    link = Link()
    lease = {"attempt_id": "a", "generation": 1, "model": "model-a", "privacy": "secret"}
    state = HostState(900, True, "normal")
    run_leased(
        config, "node-g", link, NodeMemory(), lease, [], lambda: state, lambda _s: None, lambda: 0
    )
    assert link.completed[0]["error_code"] == "privacy_refused"


def test_the_task_loop_refuses_a_lease_without_a_privacy_class(tmp_path):
    config = make_config(tmp_path, "/bin/true")
    link = TaskLink({k: v for k, v in TASK_LEASE.items() if k != "privacy"})
    task_step(config, "node-a", link, lambda: HostState(900, True, "normal"), lambda: 0.0, print)
    assert link.completed[0][2]["error_code"] == "privacy_refused"
