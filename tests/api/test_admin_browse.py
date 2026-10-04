import threading

import pytest

from tests.api.test_api_roundtrip import call
from tests.conftest import body
from worker.api.build_server import build_server
from worker.auth.add_principal import add_principal
from worker.jobs.delete_payloads import delete_payloads
from worker.jobs.submit_job import submit_job
from worker.store.migrate_state import migrate_state
from worker.store.open_store import open_store

LIST_FIELDS = {
    "id",
    "queue",
    "producer",
    "state",
    "privacy",
    "tier",
    "created",
    "updated",
    "attempts",
    "last_error",
    "acked",
    "retry_of",
    "sampling",
    "kind",
    "model",
    "priority",
    "finished",
    "deadline",
    "lease_node",
    "lease_expires",
    "parent_id",
    "preemptions",
    "tokens_in",
    "tokens_out",
    "cost_usd",
    "wall_s",
    "last_model",
    "last_provider",
    "last_started",
    "last_outcome",
}
ATTEMPT_FIELDS = {
    "node",
    "provider",
    "model",
    "outcome",
    "error",
    "started",
    "ended",
    "tokens_in",
    "tokens_out",
    "wall_s",
    "cost_usd",
}


@pytest.fixture
def admin(config, tmp_path):
    """A served coordinator, its tokens and a direct connection to its store."""
    state = tmp_path / "state"
    tokens = {
        name: add_principal(state, kind, name)
        for kind, name in (("producer", "pa"), ("admin", "admin"))
    }
    migrate_state(state)
    server = build_server(config, state)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    conn = open_store(state)
    yield f"http://127.0.0.1:{server.server_address[1]}", tokens, conn, config
    conn.close()
    server.shutdown()


def submit(conn, config, key, now, *, queue="pa.bulk", privacy="public"):
    return submit_job(conn, config, "pa", body(key, queue=queue, privacy=privacy), now)[0]


def test_every_admin_route_refuses_a_producer(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0)
    for method, path in (
        ("GET", "/v1/admin/jobs"),
        ("GET", f"/v1/admin/jobs/{job}"),
        ("GET", f"/v1/admin/jobs/{job}/content"),
        ("POST", f"/v1/admin/jobs/{job}/cancel"),
        ("POST", f"/v1/admin/jobs/{job}/retry"),
        ("POST", f"/v1/admin/jobs/{job}/ack"),
        ("GET", "/v1/admin/audit"),
    ):
        assert call(base, t["pa"], method, path) == (403, {"error": "forbidden"}), path


def test_the_list_is_newest_first_and_cursor_paged(admin):
    base, t, conn, config = admin
    ids = [submit(conn, config, f"k{i}", float(i)) for i in range(3)]
    status, page = call(base, t["admin"], "GET", "/v1/admin/jobs?limit=2")
    assert status == 200
    assert [r["id"] for r in page["jobs"]] == [ids[2], ids[1]]
    assert set(page["jobs"][0]) == LIST_FIELDS
    assert page["next"]
    status, rest = call(base, t["admin"], "GET", f"/v1/admin/jobs?limit=2&before={page['next']}")
    assert (status, [r["id"] for r in rest["jobs"]], rest["next"]) == (200, [ids[0]], None)


def test_a_row_carries_metadata_and_never_content(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 5.0, privacy="mail")
    conn.execute("UPDATE jobs SET producer='_shadow', error='timeout' WHERE id=?", (job,))
    row = call(base, t["admin"], "GET", "/v1/admin/jobs")[1]["jobs"][0]
    assert row["sampling"] is True
    assert (row["producer"], row["privacy"], row["last_error"], row["attempts"]) == (
        "_shadow",
        "mail",
        "timeout",
        0,
    )
    assert "hi" not in str(row)


def test_filters_narrow_and_bad_filters_are_refused(admin):
    base, t, conn, config = admin
    submit(conn, config, "k1", 1.0)
    live = submit(conn, config, "k2", 2.0, queue="pa.live")
    rows = call(base, t["admin"], "GET", "/v1/admin/jobs?queue=pa.live&state=queued&producer=pa")
    assert [r["id"] for r in rows[1]["jobs"]] == [live]
    assert call(base, t["admin"], "GET", "/v1/admin/jobs?producer=other")[1]["jobs"] == []
    for query, code in (
        ("queue=nope", "unknown_queue"),
        ("state=nope", "unknown_state"),
        ("before=garbage", "bad_cursor"),
        ("limit=x", "bad_request"),
    ):
        assert call(base, t["admin"], "GET", f"/v1/admin/jobs?{query}") == (400, {"error": code})


def test_limit_is_clamped(admin):
    base, t, conn, config = admin
    submit(conn, config, "k1", 1.0)
    submit(conn, config, "k2", 2.0)
    assert len(call(base, t["admin"], "GET", "/v1/admin/jobs?limit=0")[1]["jobs"]) == 1
    assert len(call(base, t["admin"], "GET", "/v1/admin/jobs?limit=5000")[1]["jobs"]) == 2


def test_the_detail_adds_attempts_and_stored_payload_flags(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0)
    conn.execute(
        "INSERT INTO attempts (id, job_id, generation, node, started, ended, outcome, error,"
        " tokens_in, tokens_out, provider, model)"
        " VALUES ('a1', ?, 1, 'node-a', 2.0, 3.0, 'failed', 'timeout', 5, 0, 'ollama', 'model-a')",
        (job,),
    )
    status, detail = call(base, t["admin"], "GET", f"/v1/admin/jobs/{job}")
    assert status == 200
    assert set(detail) == LIST_FIELDS | {"attempt_details", "results", "has_input", "has_output"}
    assert (detail["has_input"], detail["has_output"], detail["attempts"]) == (True, False, 1)
    [attempt] = detail["attempt_details"]
    assert set(attempt) == ATTEMPT_FIELDS
    assert (attempt["node"], attempt["outcome"], attempt["tokens_in"]) == ("node-a", "failed", 5)
    assert call(base, t["admin"], "GET", "/v1/admin/jobs/nope") == (404, {"error": "not_found"})


def test_public_content_needs_no_reveal_and_writes_no_audit(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0, privacy="internal")
    status, content = call(base, t["admin"], "GET", f"/v1/admin/jobs/{job}/content")
    assert status == 200
    assert content["input"]["messages"][0]["content"] == "hi"
    assert content["output"] is None
    assert conn.execute("SELECT count(*) FROM audit").fetchone()[0] == 0


def test_sensitive_content_needs_the_matching_reveal_header(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0, privacy="mail")
    path = f"/v1/admin/jobs/{job}/content"
    refused = (403, {"error": "reveal_required"})
    assert call(base, t["admin"], "GET", path) == refused
    assert call(base, t["admin"], "GET", path, headers={"X-Worker-Reveal": "personal"}) == refused
    assert conn.execute("SELECT count(*) FROM audit").fetchone()[0] == 0
    status, content = call(base, t["admin"], "GET", path, headers={"X-Worker-Reveal": "mail"})
    assert (status, content["input"]["messages"][0]["content"]) == (200, "hi")
    row = conn.execute("SELECT * FROM audit").fetchone()
    assert (row["action"], row["job_id"], row["privacy"], row["principal"]) == (
        "reveal",
        job,
        "mail",
        "admin",
    )


def test_content_whose_payloads_are_gone_is_410(admin):
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0)
    delete_payloads(conn, job)
    assert call(base, t["admin"], "GET", f"/v1/admin/jobs/{job}/content") == (
        410,
        {"error": "content_gone"},
    )
