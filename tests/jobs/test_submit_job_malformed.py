import json

import pytest

from tests.conftest import body
from worker.jobs.api_error import ApiError
from worker.jobs.submit_job import submit_job


@pytest.mark.parametrize(
    ("raw", "status", "code"),
    [
        (b"\xff\xfe not json", 400, "bad_json"),
        (b"{broken", 400, "bad_json"),
        (json.dumps([1, 2]).encode(), 400, "bad_json"),
        (body(priority="abc"), 400, "out_of_range"),
        (body(priority=True), 400, "out_of_range"),
        (body(priority=None), 400, "out_of_range"),
        (body(max_attempts=False), 400, "out_of_range"),
        (body(deadline="soon"), 400, "out_of_range"),
        (body(requirements=["model-a"]), 400, "missing_models"),
        (body(parent_id=7), 400, "bad_split_child"),
    ],
)
def test_malformed_input_is_an_api_error(conn, config, raw, status, code):
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", raw, 100.0)
    assert (error.value.status, error.value.code) == (status, code)


def test_cross_queue_child_is_refused_and_parent_stays(conn, config):
    parent, _ = submit_job(conn, config, "pa", body(key="p"), 100.0)
    conn.execute("UPDATE jobs SET state='split_requested' WHERE id=?", (parent,))
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(key="p/0/2", queue="pa.live", parent_id=parent), 101.0)
    assert (error.value.status, error.value.code) == (400, "bad_split_child")
    state = conn.execute("SELECT state FROM jobs WHERE id=?", (parent,)).fetchone()["state"]
    assert state == "split_requested"
