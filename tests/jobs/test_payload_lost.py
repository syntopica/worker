import pytest

from tests.conftest import body
from worker.jobs.api_error import ApiError
from worker.jobs.candidate import Candidate
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.list_results import list_results
from worker.jobs.reconcile_lost_payloads import reconcile_lost_payloads
from worker.jobs.submit_job import submit_job
from worker.policy.candidate_eligible import candidate_eligible
from worker.policy.pick_job import pick_job

ASK = LeaseRequest("node-a", None, False, 44.0, 900.0)
OK = CompletionReport("succeeded", {"text": "ok", "json": None}, {}, {"node": "node-a"}, None, 1.0)


def state_of(conn, job_id):
    row = conn.execute("SELECT state, error FROM jobs WHERE id=?", (job_id,)).fetchone()
    return row["state"], row["error"]


def test_a_job_without_input_fails_and_the_lease_picks_the_next(conn, config):
    lost, _ = submit_job(conn, config, "pa", body(key="lost", queue="pa.live", priority=90), 0.0)
    good, _ = submit_job(conn, config, "pa", body(key="good", queue="pa.live"), 0.0)
    conn.execute("DELETE FROM p.inputs WHERE job_id=?", (lost,))
    lease = lease_job(conn, config, ASK, 1.0)
    assert lease.job_id == good
    assert state_of(conn, lost) == ("failed", "payload_lost")
    controls = [r["control"] for r in list_results(conn, "pa", "pa.live", 0, 10)]
    assert controls == ["failed"]


def test_only_lost_jobs_leave_nothing_to_lease(conn, config):
    for i in range(3):
        submit_job(conn, config, "pa", body(key=f"k{i}", queue="pa.live"), 0.0)
    conn.execute("DELETE FROM p.inputs")
    assert lease_job(conn, config, ASK, 1.0) is None
    assert {r[0] for r in conn.execute("SELECT error FROM jobs")} == {"payload_lost"}


def test_completing_a_job_whose_input_vanished_fails_it_as_payload_lost(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    lease = lease_job(conn, config, ASK, 1.0)
    conn.execute("DELETE FROM p.inputs")
    assert complete_attempt(conn, config, lease.attempt_id, lease.generation, OK, 2.0) == "failed"
    assert state_of(conn, job_id) == ("failed", "payload_lost")


def test_restore_fails_every_live_job_whose_payload_is_missing(conn, config):
    lost, _ = submit_job(conn, config, "pa", body(key="lost", queue="pa.live"), 0.0)
    kept, _ = submit_job(conn, config, "pa", body(key="kept", queue="pa.live"), 0.0)
    done, _ = submit_job(conn, config, "pa", body(key="done", queue="pa.live"), 0.0)
    conn.execute("UPDATE jobs SET state='cancelled' WHERE id=?", (done,))
    conn.execute("DELETE FROM p.inputs WHERE job_id IN (?, ?)", (lost, done))
    assert reconcile_lost_payloads(conn, 5.0) == 1
    assert state_of(conn, lost) == ("failed", "payload_lost")
    assert state_of(conn, kept) == ("queued", None)
    assert state_of(conn, done) == ("cancelled", None)


def test_an_unknown_queue_is_ineligible_and_never_raises(conn, config):
    stray = Candidate("j", "gone.queue", "model-a", 50, "public", 0.0, False, 0.0)
    assert candidate_eligible(stray, ASK, config) is False
    assert pick_job([stray], ASK, config, {}, 1.0) is None
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    conn.execute("UPDATE jobs SET queue='gone.queue' WHERE id=?", (job_id,))
    assert lease_job(conn, config, ASK, 1.0) is None


@pytest.mark.parametrize(
    "shape",
    [
        {},
        {"messages": []},
        {"messages": "hi"},
        {"messages": ["hi"]},
        {"messages": [{"role": "user"}]},
        {"messages": [{"role": 1, "content": "hi"}]},
        {"messages": [{"role": "user", "content": ["hi"]}]},
        {"messages": [{"role": "user", "content": "hi"}], "options": []},
        {"messages": [{"role": "user", "content": "hi"}], "schema": []},
        {"messages": [{"role": "user", "content": "hi"}], "format": "xml"},
    ],
)
def test_a_malformed_inference_input_is_bad_input(conn, config, shape):
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(input=shape), 0.0)
    assert (error.value.status, error.value.code) == (400, "bad_input")


def test_a_well_formed_input_with_every_optional_field_is_accepted(conn, config):
    shape = {
        "messages": [{"role": "user", "content": "hi"}],
        "options": {"temperature": 0},
        "schema": {"type": "object"},
        "format": "json",
    }
    assert submit_job(conn, config, "pa", body(input=shape), 0.0)[1]


@pytest.mark.parametrize("key", ["k0/01/2", "k0/0/02", "other/0/2", "k0x/0/2"])
def test_a_child_key_must_be_exactly_parent_index_count(conn, config, key):
    parent, _ = submit_job(conn, config, "pa", body(key="k0"), 0.0)
    conn.execute("UPDATE jobs SET state='split_requested' WHERE id=?", (parent,))
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(key=key, parent_id=parent), 1.0)
    assert error.value.code in ("bad_split_child", "bad_split_count")
    assert submit_job(conn, config, "pa", body(key="k0/1/2", parent_id=parent), 1.0)[1]
