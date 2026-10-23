import json
from contextlib import closing

import pytest

from tests.conftest import body, fresh_store
from tests.ladder.test_on_demand_profile import make_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.insert_shadow_job import insert_shadow_job
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.read_audit import read_audit
from worker.jobs.submit_job import submit_job


@pytest.mark.parametrize("pin", ["ollama:model-a", "openrouter:vendor/model:free", "runner:other"])
def test_an_explicit_override_keeps_the_pin_and_records_the_actual_executor(tmp_path, pin):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        source, _ = submit_job(conn, config, "pa", body(privacy="public"), 0.0)
        row = conn.execute("SELECT * FROM jobs WHERE id=?", (source,)).fetchone()
        conn.execute("UPDATE jobs SET state='cancelled' WHERE id=?", (source,))
        member = insert_shadow_job(
            conn,
            row,
            {"messages": [{"role": "user", "content": "hi"}]},
            group=source,
            pin=pin,
            model="model-a",
            now=0.0,
        )
        protected = LeaseRequest("node-a", None, False, 0.0, 0.0, "task", "max")
        assert lease_job(conn, config, protected, 1.0) is None
        override = LeaseRequest("node-a", None, False, 0.0, 0.0, "task", "max", override_pins=True)
        got = lease_job(conn, config, override, 1.0)
        assert got.job_id == member
        assert conn.execute("SELECT pin FROM jobs WHERE id=?", (member,)).fetchone()[0] == pin
        executor = {"provider": "max-lane", "model": "claude-x"}
        report = CompletionReport(
            "succeeded", {"text": "OK", "json": None}, {}, executor, None, 1.0
        )
        assert (
            complete_attempt(conn, config, got.attempt_id, got.generation, report, 2.0)
            == "succeeded"
        )
        actual = json.loads(
            conn.execute("SELECT executor FROM results WHERE job_id=?", (member,)).fetchone()[0]
        )
        assert actual == executor
        assert read_audit(conn, 0.0)[0]["action"] == "override_pin"


def test_an_override_cannot_run_an_original_answer_marker(tmp_path):
    config = make_config(tmp_path)
    with closing(fresh_store(tmp_path / "state")) as conn:
        source, _ = submit_job(conn, config, "pa", body(privacy="public"), 0.0)
        conn.execute("UPDATE jobs SET pin='original' WHERE id=?", (source,))
        req = LeaseRequest("node-a", None, False, 0.0, 0.0, "task", "max", override_pins=True)
        assert lease_job(conn, config, req, 1.0) is None
