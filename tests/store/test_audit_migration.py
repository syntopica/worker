import sqlite3

from tests.conftest import fresh_store
from worker.store.schema_sql import SCHEMA
from worker.store.store_version import STORE_VERSION

AUDIT_START = SCHEMA.index("CREATE TABLE IF NOT EXISTS audit")
OLD = (SCHEMA[:AUDIT_START] + SCHEMA[SCHEMA.index(";", AUDIT_START) + 1 :]).replace(
    "  pin TEXT,\n  retry_of TEXT,\n", "  pin TEXT,\n"
)


def test_a_version_7_store_gains_the_audit_table_and_retry_of(tmp_path):
    assert "retry_of" not in OLD and "audit" not in OLD
    state = tmp_path / "state"
    state.mkdir()
    conn = sqlite3.connect(state / "meta.sqlite3")
    conn.execute("ATTACH DATABASE ? AS p", (str(state / "payloads.sqlite3"),))
    conn.executescript(OLD)
    conn.execute("PRAGMA user_version=7")
    conn.commit()
    conn.close()
    store = fresh_store(state)
    assert STORE_VERSION == 8
    assert store.execute("PRAGMA user_version").fetchone()[0] == 8
    assert "retry_of" in {r[1] for r in store.execute("PRAGMA table_info(jobs)")}
    columns = {r[1] for r in store.execute("PRAGMA table_info(audit)")}
    assert columns == {"seq", "action", "job_id", "privacy", "principal", "created"}
