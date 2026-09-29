import pytest

from tests.conftest import body
from worker.jobs.ack_result import ack_result
from worker.jobs.api_error import ApiError
from worker.jobs.cancel_job import cancel_job
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.get_job import get_job
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.list_results import list_results
from worker.jobs.read_status import read_status
from worker.jobs.submit_job import submit_job
from worker.jobs.sweep_retention import sweep_retention

DAY = 86400.0


def succeed(conn, config, key="k1", privacy="mail", now=0.0):
    job_id, _ = submit_job(conn, config, "pa", body(key=key, privacy=privacy), now)
    lease = lease_job(conn, config, LeaseRequest("node-a", None, False, 44.0, 9999.0), now + 1)
    report = CompletionReport(
        "succeeded", {"text": "ok", "json": None}, {}, {"node": "node-a"}, None, 1.0
    )
    complete_attempt(conn, config, lease.attempt_id, lease.generation, report, now + 2)
    return job_id


def test_results_feed_is_ordered_and_ack_deletes_sensitive_payloads(conn, config):
    job_id = succeed(conn, config)
    results = list_results(conn, "pa", "pa.bulk", 0, 10)
    assert [r["output"]["text"] for r in results] == ["ok"]
    ack_result(conn, config, "pa", job_id, results[0]["result_id"], False, 10.0)
    assert list_results(conn, "pa", "pa.bulk", 0, 10) == []
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 0
    assert conn.execute("SELECT count(*) FROM p.outputs").fetchone()[0] == 0


def test_another_producer_cannot_read_the_job(conn, config):
    job_id = succeed(conn, config)
    with pytest.raises(ApiError):
        get_job(conn, "other", job_id)


def test_declining_a_split_parks_the_job_with_a_minimum_idle(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    conn.execute("UPDATE jobs SET state='split_requested'")
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, control, created) VALUES ('r1', ?, 'pa', 'pa.bulk', 'split_requested', 1)",
        (job_id,),
    )
    ack_result(conn, config, "pa", job_id, "r1", True, 5.0)
    row = conn.execute("SELECT state, parked, parked_min_idle_s, acked FROM jobs").fetchone()
    assert (row["state"], row["parked"], row["parked_min_idle_s"], row["acked"]) == (
        "queued",
        1,
        600.0,
        None,
    )


def test_unacknowledged_sensitive_result_expires_after_72_hours(conn, config):
    succeed(conn, config)
    sweep_retention(conn, config, 71 * 3600.0)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "succeeded"
    sweep_retention(conn, config, 73 * 3600.0)
    assert conn.execute("SELECT state FROM jobs").fetchone()[0] == "unacked_expired"
    assert conn.execute("SELECT count(*) FROM p.outputs").fetchone()[0] == 0
    assert [r["control"] for r in list_results(conn, "pa", "pa.bulk", 0, 10)][
        -1
    ] == "unacked_expired"


def test_public_payload_kept_for_retention_after_ack(conn, config):
    job_id = succeed(conn, config, privacy="public")
    result = list_results(conn, "pa", "pa.bulk", 0, 10)[0]
    ack_result(conn, config, "pa", job_id, result["result_id"], False, 10.0)
    sweep_retention(conn, config, 6 * DAY)
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 1
    sweep_retention(conn, config, 8 * DAY)
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 0


def test_cancel_a_queued_job_and_status_counts_it(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    assert cancel_job(conn, "pa", job_id, 1.0) == "cancelled"
    status = read_status(conn, 2.0)
    assert status["queues"]["pa.bulk"]["states"]["cancelled"] == 1


def test_get_job_finds_its_result_behind_a_hundred_others(conn, config):
    other = succeed(conn, config, key="other", now=0.0)
    for i in range(101):
        conn.execute(
            "INSERT INTO results (result_id, job_id, producer, queue, control, created)"
            " VALUES (?, ?, 'pa', 'pa.bulk', 'failed', ?)",
            (f"f{i}", other, float(i)),
        )
    target = succeed(conn, config, key="target", now=2000.0)
    assert get_job(conn, "pa", target)["result"]["job_id"] == target


def test_unacked_public_payload_is_retained_from_finish(conn, config):
    succeed(conn, config, privacy="public")
    sweep_retention(conn, config, 6 * DAY)
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 1
    sweep_retention(conn, config, 8 * DAY)
    assert conn.execute("SELECT count(*) FROM p.inputs").fetchone()[0] == 0


def test_duplicate_ack_keeps_the_first_time(conn, config):
    job_id = succeed(conn, config, privacy="public")
    rid = list_results(conn, "pa", "pa.bulk", 0, 10)[0]["result_id"]
    ack_result(conn, config, "pa", job_id, rid, False, 10.0)
    ack_result(conn, config, "pa", job_id, rid, False, 50.0)
    assert conn.execute("SELECT acked FROM jobs").fetchone()[0] == 10.0
    assert conn.execute("SELECT acked FROM results").fetchone()[0] == 10.0
