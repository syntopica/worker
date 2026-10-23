import sqlite3
from contextlib import closing

from tests.conftest import fresh_store
from worker.store.schema_sql import SCHEMA
from worker.store.store_version import STORE_VERSION


def test_version_8_attempts_keep_their_metadata_and_gain_a_nullable_profile(tmp_path):
    state = tmp_path / "state"
    state.mkdir()
    with closing(sqlite3.connect(state / "meta.sqlite3")) as conn:
        conn.execute("ATTACH DATABASE ? AS p", (str(state / "payloads.sqlite3"),))
        conn.executescript(SCHEMA)
        conn.execute(
            "INSERT INTO attempts (id,job_id,generation,node,started,provider,model) VALUES ('a','j',1,'node-a',0,'codex','model-x')"
        )
        conn.execute("PRAGMA user_version=8")
        conn.commit()
    for _ in range(2):
        with closing(fresh_store(state)) as conn:
            row = conn.execute(
                "SELECT provider,model,profile FROM attempts WHERE id='a'"
            ).fetchone()
            assert tuple(row) == ("codex", "model-x", None)
            assert conn.execute("PRAGMA user_version").fetchone()[0] == STORE_VERSION
