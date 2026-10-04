import json

from tests.api.test_admin_browse import admin, submit  # noqa: F401
from tests.api.test_api_roundtrip import call

COLUMNS = ("outcome", "tokens_in", "tokens_out", "wall_s", "cost_usd", "provider", "model")


def add_attempt(conn, job, attempt, started, **values):
    conn.execute(
        "INSERT INTO attempts (id, job_id, started, node, generation, outcome, tokens_in,"
        " tokens_out, wall_s, cost_usd, provider, model)"
        " VALUES (?, ?, ?, 'node-a', 1, ?, ?, ?, ?, ?, ?, ?)",
        (attempt, job, started, *(values.get(c) for c in COLUMNS)),
    )


def test_running_now_is_a_comma_separated_state_filter(admin):  # noqa: F811
    base, t, conn, config = admin
    queued = submit(conn, config, "k1", 1.0)
    leased = submit(conn, config, "k2", 2.0)
    running = submit(conn, config, "k3", 3.0)
    conn.execute("UPDATE jobs SET state='leased', lease_node='node-a' WHERE id=?", (leased,))
    conn.execute("UPDATE jobs SET state='running', lease_expires=99.0 WHERE id=?", (running,))
    status, page = call(base, t["admin"], "GET", "/v1/admin/jobs?state=leased,running,draining")
    assert status == 200
    assert [r["id"] for r in page["jobs"]] == [running, leased]
    assert queued not in {r["id"] for r in page["jobs"]}
    assert (page["jobs"][1]["lease_node"], page["jobs"][0]["lease_expires"]) == ("node-a", 99.0)
    assert call(base, t["admin"], "GET", "/v1/admin/jobs?state=running,nope") == (
        400,
        {"error": "unknown_state"},
    )


def test_a_row_sums_attempt_usage_and_names_the_newest_attempt(admin):  # noqa: F811
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0)
    add_attempt(conn, job, "a1", 2.0, outcome="failed", tokens_in=5, tokens_out=1, wall_s=3.0)
    add_attempt(
        conn,
        job,
        "a2",
        4.0,
        outcome="succeeded",
        tokens_in=7,
        tokens_out=2,
        wall_s=1.5,
        cost_usd=0.25,
        provider="openrouter",
        model="model-b",
    )
    [row] = call(base, t["admin"], "GET", "/v1/admin/jobs")[1]["jobs"]
    assert (row["kind"], row["model"], row["priority"], row["preemptions"]) == (
        "inference",
        "model-a",
        50,
        0,
    )
    assert (row["tokens_in"], row["tokens_out"], row["wall_s"], row["cost_usd"]) == (
        12,
        3,
        4.5,
        0.25,
    )
    assert (row["last_model"], row["last_provider"], row["last_started"]) == (
        "model-b",
        "openrouter",
        4.0,
    )
    assert row["last_outcome"] == "succeeded"


def test_a_row_without_attempts_has_null_totals(admin):  # noqa: F811
    base, t, conn, config = admin
    submit(conn, config, "k1", 1.0)
    [row] = call(base, t["admin"], "GET", "/v1/admin/jobs")[1]["jobs"]
    assert (row["attempts"], row["tokens_in"], row["cost_usd"], row["last_model"]) == (
        0,
        None,
        None,
        None,
    )


def test_the_detail_adds_attempt_wall_and_cost_and_result_metadata(admin):  # noqa: F811
    base, t, conn, config = admin
    job = submit(conn, config, "k1", 1.0)
    add_attempt(conn, job, "a1", 2.0, outcome="succeeded", wall_s=2.5, cost_usd=0.1)
    executor = {"node": "node-a", "provider": "ollama", "model": "model-a"}
    usage = {"tokens_in": 4, "tokens_out": 2}
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, executor, usage, created, rating)"
        " VALUES ('r1', ?, 'pa', 'pa.bulk', ?, ?, 3.0, 'good')",
        (job, json.dumps(executor), json.dumps(usage)),
    )
    conn.execute("INSERT INTO p.outputs (result_id, body) VALUES ('r1', ?)", ('{"text": "x"}',))
    conn.execute(
        "INSERT INTO results (result_id, job_id, producer, queue, control, detail, created)"
        " VALUES ('r0', ?, 'pa', 'pa.bulk', 'failed', ?, 2.5)",
        (job, json.dumps({"error": "schema_violation", "schema_path": "/a"})),
    )
    detail = call(base, t["admin"], "GET", f"/v1/admin/jobs/{job}")[1]
    [attempt] = detail["attempt_details"]
    assert (attempt["wall_s"], attempt["cost_usd"]) == (2.5, 0.1)
    first, second = detail["results"]
    assert (first["result_id"], first["executor"], first["usage"], first["rating"]) == (
        "r1",
        executor,
        usage,
        "good",
    )
    assert (second["control"], second["detail"]["schema_path"]) == ("failed", "/a")
    assert "output" not in first
    assert '"x"' not in json.dumps(detail["results"])
