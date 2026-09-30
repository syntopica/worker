import dataclasses

from tests.conftest import body
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job


def test_a_removed_queue_backlog_does_not_starve_configured_queues(conn, config):
    old = dataclasses.replace(config.queues["pa.bulk"], name="pa.old", max_outstanding=600)
    wide = dataclasses.replace(
        config,
        queues={**config.queues, "pa.old": old},
        producers={"pa": frozenset({*config.producers["pa"], "pa.old"})},
    )
    for i in range(510):
        submit_job(conn, wide, "pa", body(key=f"old{i}", queue="pa.old", priority=90), 0.0)
    live_id, _ = submit_job(conn, config, "pa", body(key="live", queue="pa.live"), 0.0)
    lease = lease_job(conn, config, LeaseRequest("node-a", None, True, 44.0, 900.0), 1.0)
    assert lease is not None
    assert lease.job_id == live_id
