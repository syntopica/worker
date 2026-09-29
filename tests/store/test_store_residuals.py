import sqlite3
import time

import pytest

from tests.conftest import fresh_store
from tests.store.test_store_fixes import legacy_store
from worker.store import connect_store as connect_module
from worker.store.open_store import open_store
from worker.store.store_not_migrated_error import StoreNotMigratedError
from worker.store.store_version import STORE_VERSION


def test_an_unmigrated_store_is_refused(tmp_path):
    with pytest.raises(StoreNotMigratedError):
        open_store(tmp_path / "new")
    legacy_store(tmp_path / "old")
    with pytest.raises(StoreNotMigratedError):
        open_store(tmp_path / "old")
    assert fresh_store(tmp_path / "old").execute("PRAGMA user_version").fetchone()[0] == (
        STORE_VERSION
    )


def test_open_store_begins_nothing_and_changes_no_schema(tmp_path, monkeypatch):
    fresh_store(tmp_path / "state").close()
    statements = []
    real = sqlite3.connect

    def traced(*args, **kwargs):
        conn = real(*args, **kwargs)
        conn.set_trace_callback(statements.append)
        return conn

    monkeypatch.setattr(connect_module.sqlite3, "connect", traced)
    open_store(tmp_path / "state").close()
    assert statements
    words = ("BEGIN", "CREATE", "ALTER", "COMMIT", "INSERT", "UPDATE", "DELETE")
    assert [s for s in statements if s.lstrip().upper().startswith(words)] == []


def test_a_read_completes_while_another_connection_holds_the_writer_lock(tmp_path):
    state = tmp_path / "state"
    writer = fresh_store(state)
    writer.execute("BEGIN IMMEDIATE")
    try:
        started = time.monotonic()
        reader = open_store(state)
        assert reader.execute("SELECT count(*) FROM jobs").fetchone()[0] == 0
        assert time.monotonic() - started < 2.0  # the busy timeout is 10 s
    finally:
        writer.execute("ROLLBACK")
