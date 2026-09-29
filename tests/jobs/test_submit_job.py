import pytest

from tests.conftest import body
from worker.jobs.api_error import ApiError
from worker.jobs.submit_job import submit_job


def test_same_key_same_payload_returns_the_same_job(conn, config):
    first, created = submit_job(conn, config, "pa", body(), 100.0)
    again, created_again = submit_job(conn, config, "pa", body(), 101.0)
    assert (again, created, created_again) == (first, True, False)


def test_same_key_different_payload_is_a_conflict(conn, config):
    submit_job(conn, config, "pa", body(), 100.0)
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(priority=90), 101.0)
    assert (error.value.status, error.value.code) == (409, "idempotency_conflict")


def test_payload_lives_only_in_the_payload_file(conn, config):
    job_id, _ = submit_job(conn, config, "pa", body(), 100.0)
    assert (
        conn.execute("SELECT count(*) FROM p.inputs WHERE job_id=?", (job_id,)).fetchone()[0] == 1
    )
    row = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
    assert "hi" not in " ".join(str(v) for v in tuple(row))


@pytest.mark.parametrize(
    ("raw", "status", "code"),
    [
        (body(contract=2), 400, "unsupported_contract"),
        (body(kind="task"), 400, "unsupported_kind"),
        (body(privacy="nope"), 400, "unknown_privacy"),
        (body(queue="other.q"), 403, "queue_not_granted"),
        (body(requirements={"capability": "chat", "models": ["missing"]}), 400, "unknown_model"),
    ],
)
def test_invalid_requests_are_refused_with_a_code(conn, config, raw, status, code):
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", raw, 100.0)
    assert (error.value.status, error.value.code) == (status, code)


def test_oversized_request_is_413(conn, config):
    big = body(input={"messages": [{"role": "user", "content": "x" * 1_100_000}]})
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", big, 100.0)
    assert error.value.status == 413


def test_outstanding_limit_returns_429(conn, config):
    for i in range(3):
        submit_job(conn, config, "pa", body(key=f"k{i}"), 100.0)
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(key="k9"), 100.0)
    assert error.value.status == 429


def test_split_children_bypass_the_limit_and_supersede_the_parent(conn, config):
    ids = [submit_job(conn, config, "pa", body(key=f"k{i}"), 100.0)[0] for i in range(3)]
    conn.execute("UPDATE jobs SET state='split_requested' WHERE id=?", (ids[0],))
    child, _ = submit_job(conn, config, "pa", body(key="k0/0/2", parent_id=ids[0]), 101.0)
    parent = conn.execute("SELECT state, split_count FROM jobs WHERE id=?", (ids[0],)).fetchone()
    assert (parent["state"], parent["split_count"]) == ("superseded", 2)
    with pytest.raises(ApiError) as error:
        submit_job(conn, config, "pa", body(key="k0/1/3", parent_id=ids[0]), 101.0)
    assert error.value.code == "split_count_mismatch"
    assert child
