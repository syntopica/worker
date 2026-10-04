import json

from worker.store.open_store import open_store


def test_audit_lists_rows_as_json_and_as_text(served, run):
    assert run("token", "add", "--kind", "admin", "--name", "admin")[0] == 0
    code, out, _ = run("audit", "--days", "1", "--json")
    assert (code, json.loads(out)) == (0, [])
    conn = open_store(served / "worker" / "state")
    conn.execute(
        "INSERT INTO audit (action, job_id, privacy, principal, created)"
        " VALUES ('reveal', 'job-1', 'mail', 'admin', strftime('%s', 'now'))"
    )
    conn.close()
    code, out, _ = run("audit", "--days", "1", "--json")
    rows = json.loads(out)
    assert code == 0
    assert [(r["action"], r["job_id"], r["class"], r["principal"]) for r in rows] == [
        ("reveal", "job-1", "mail", "admin")
    ]
    code, out, _ = run("audit", "--days", "1")
    assert code == 0
    assert "reveal" in out and "job-1" in out and "mail" in out
