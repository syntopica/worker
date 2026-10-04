from tests.api.test_admin_browse import admin, submit  # noqa: F401
from tests.api.test_api_roundtrip import call
from worker.jobs.cancel_job import cancel_job
from worker.jobs.finish_failed import finish_failed
from worker.jobs.load_job_input import load_job_input


def fail(conn, job_id, now=3.0):
    row = conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()
    finish_failed(conn, row, "timeout", now)


def audit_rows(conn):
    return [
        (r["action"], r["job_id"], r["privacy"], r["principal"])
        for r in conn.execute("SELECT * FROM audit ORDER BY seq")
    ]


def job(conn, job_id):
    return conn.execute("SELECT * FROM jobs WHERE id=?", (job_id,)).fetchone()


def test_admin_cancels_any_producers_job_and_audits_it(admin):  # noqa: F811
    base, t, conn, config = admin
    queued = submit(conn, config, "k1", 1.0, privacy="mail")
    path = f"/v1/admin/jobs/{queued}/cancel"
    assert call(base, t["admin"], "POST", path) == (200, {"id": queued, "state": "cancelled"})
    assert call(base, t["admin"], "POST", path) == (200, {"id": queued, "state": "cancelled"})
    assert audit_rows(conn)[0] == ("cancel", queued, "mail", "admin")
    assert call(base, t["admin"], "POST", "/v1/admin/jobs/nope/cancel") == (
        404,
        {"error": "not_found"},
    )


def test_retry_admits_a_copy_and_acknowledges_the_original(admin):  # noqa: F811
    base, t, conn, config = admin
    original = submit(conn, config, "k1", 1.0, privacy="internal")
    fail(conn, original)
    status, answer = call(base, t["admin"], "POST", f"/v1/admin/jobs/{original}/retry")
    assert status == 201
    new = answer["id"]
    assert answer == {"id": new, "state": "queued", "retry_of": original}
    copy, old = job(conn, new), job(conn, original)
    for field in ("queue", "producer", "privacy", "tier", "model", "kind", "priority"):
        assert copy[field] == old[field], field
    assert (copy["retry_of"], copy["deadline"]) == (original, None)
    assert load_job_input(conn, new) == load_job_input(conn, original)
    assert old["acked"] is not None
    unacked = "SELECT count(*) FROM results WHERE job_id=? AND acked IS NULL"
    assert conn.execute(unacked, (original,)).fetchone()[0] == 0
    assert audit_rows(conn) == [("retry", original, "internal", "admin")]
    again = call(base, t["admin"], "POST", f"/v1/admin/jobs/{original}/retry")
    assert again == (200, {"id": new, "state": "queued", "retry_of": original})


def test_retrying_a_sensitive_job_keeps_its_class_and_drops_the_old_payload(admin):  # noqa: F811
    base, t, conn, config = admin
    original = submit(conn, config, "k1", 1.0, privacy="secret")
    fail(conn, original)
    status, answer = call(base, t["admin"], "POST", f"/v1/admin/jobs/{original}/retry")
    assert status == 201
    assert job(conn, answer["id"])["privacy"] == "secret"
    assert load_job_input(conn, original) is None
    assert load_job_input(conn, answer["id"]) is not None


def test_only_finished_jobs_with_their_input_can_be_retried(admin):  # noqa: F811
    base, t, conn, config = admin
    queued = submit(conn, config, "k1", 1.0)
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{queued}/retry") == (
        409,
        {"error": "not_retryable"},
    )
    gone = submit(conn, config, "k2", 2.0, privacy="mail")
    cancel_job(conn, "pa", gone, 3.0)  # a sensitive cancel deletes the payloads
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{gone}/retry") == (
        410,
        {"error": "content_gone"},
    )
    assert audit_rows(conn) == []


def test_retry_respects_the_outstanding_limit_and_changes_nothing(admin):  # noqa: F811
    base, t, conn, config = admin
    original = submit(conn, config, "k0", 1.0)
    cancel_job(conn, "pa", original, 2.0)  # public: the input stays stored
    for i in range(3):
        submit(conn, config, f"k{i + 1}", 3.0 + i)  # pa.bulk allows 3 outstanding
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{original}/retry") == (
        429,
        {"error": "outstanding_limit"},
    )
    assert conn.execute("SELECT count(*) FROM jobs").fetchone()[0] == 4
    assert audit_rows(conn) == []


def test_retry_is_admitted_as_a_submit_so_sampling_jobs_are_refused(admin):  # noqa: F811
    base, t, conn, config = admin
    original = submit(conn, config, "k1", 1.0)
    fail(conn, original)
    conn.execute("UPDATE jobs SET producer='_shadow' WHERE id=?", (original,))
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{original}/retry") == (
        403,
        {"error": "queue_not_granted"},
    )
    assert job(conn, original)["acked"] is None


def test_ack_clears_a_failed_job_from_the_outstanding_count(admin):  # noqa: F811
    base, t, conn, config = admin
    failed = submit(conn, config, "k1", 1.0, privacy="mail")
    fail(conn, failed)
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{failed}/ack") == (
        200,
        {"id": failed, "state": "failed"},
    )
    assert job(conn, failed)["acked"] is not None
    assert load_job_input(conn, failed) is None
    assert audit_rows(conn) == [("ack", failed, "mail", "admin")]


def test_ack_refuses_a_job_without_a_control_result(admin):  # noqa: F811
    base, t, conn, config = admin
    queued = submit(conn, config, "k1", 1.0)
    refused = (409, {"error": "not_ackable"})
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{queued}/ack") == refused
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, created)"
        " VALUES ('r1', ?, 'pa', 'pa.bulk', 2.0)",
        (queued,),
    )
    assert call(base, t["admin"], "POST", f"/v1/admin/jobs/{queued}/ack") == refused


def test_the_audit_route_lists_rows_newest_first(admin):  # noqa: F811
    base, t, conn, config = admin
    first = submit(conn, config, "k1", 1.0)
    second = submit(conn, config, "k2", 2.0)
    call(base, t["admin"], "POST", f"/v1/admin/jobs/{first}/cancel")
    call(base, t["admin"], "POST", f"/v1/admin/jobs/{second}/cancel")
    status, payload = call(base, t["admin"], "GET", "/v1/admin/audit?days=1")
    assert status == 200
    assert [r["job_id"] for r in payload["rows"]] == [second, first]
    assert set(payload["rows"][0]) == {"time", "action", "job_id", "class", "principal"}
