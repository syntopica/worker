import sqlite3
import stat
import subprocess
import sys

import pytest

from worker.store import exclude_payloads_from_backup as exclude_module
from worker.store import mark_backup_excluded as mark_module
from worker.store.open_store import open_store
from worker.store.schema_sql import SCHEMA

LEGACY = SCHEMA.replace("  payloads_deleted INTEGER NOT NULL DEFAULT 0,\n", "")


def legacy_store(state):
    state.mkdir()
    conn = sqlite3.connect(state / "meta.sqlite3")
    conn.execute("ATTACH DATABASE ? AS p", (str(state / "payloads.sqlite3"),))
    conn.executescript(LEGACY)
    conn.execute(
        "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy,"
        " model, state, max_attempts, not_before, created, updated, finished)"
        " VALUES ('old', 'pa', 'pa.bulk', 'inference', 'k', 'h', 50, 'mail', 'model-a', 'failed', 3,"
        " 0, 0, 0, 1)"
    )
    conn.execute(
        "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy,"
        " model, state, max_attempts, not_before, created, updated)"
        " VALUES ('live', 'pa', 'pa.bulk', 'inference', 'k2', 'h', 50, 'mail', 'model-a', 'queued', 3,"
        " 0, 0, 0)"
    )
    conn.execute("INSERT INTO p.inputs (job_id, body) VALUES ('live', '{}')")
    conn.commit()
    conn.close()


def test_the_legacy_store_gains_the_column_and_keeps_its_rows(tmp_path):
    assert "payloads_deleted" not in LEGACY
    legacy_store(tmp_path / "state")
    conn = open_store(tmp_path / "state")
    rows = dict(conn.execute("SELECT id, payloads_deleted FROM jobs").fetchall())
    assert rows == {"old": 1, "live": 0}
    conn.close()
    conn = open_store(tmp_path / "state")  # idempotent on the second open
    assert conn.execute("SELECT count(*) FROM jobs").fetchone()[0] == 2


def test_the_hot_path_indexes_exist(conn):
    names = {r[0] for r in conn.execute("SELECT name FROM sqlite_master WHERE type='index'")}
    assert {"results_job", "results_feed", "jobs_outstanding", "jobs_payloads"} <= names


def test_state_files_are_private(tmp_path):
    state = tmp_path / "state"
    conn = open_store(state)
    conn.execute("INSERT INTO nodes (name, report, updated) VALUES ('n', '{}', 0)")
    conn.close()
    conn = open_store(state)
    assert stat.S_IMODE(state.stat().st_mode) == 0o700
    for path in state.iterdir():
        assert stat.S_IMODE(path.stat().st_mode) == 0o600, path.name


def test_the_payload_wal_is_truncated_after_checkpoints(conn):
    assert conn.execute("PRAGMA p.journal_size_limit").fetchone()[0] == 0


@pytest.mark.skipif(sys.platform != "darwin", reason="Time Machine exists on macOS only")
def test_payload_files_are_excluded_from_time_machine(tmp_path):
    state = tmp_path / "state"
    open_store(state)
    value = subprocess.run(
        [
            "/usr/bin/xattr",
            "-p",
            "com.apple.metadata:com_apple_backup_excludeItem",
            str(state / "payloads.sqlite3"),
        ],
        capture_output=True,
        check=False,
    )
    assert value.returncode == 0
    assert b"com.apple.backupd" in value.stdout
    meta = subprocess.run(
        ["/usr/bin/xattr", str(state / "meta.sqlite3")], capture_output=True, check=False
    )
    assert b"backup_excludeItem" not in meta.stdout


def test_a_failing_exclusion_logs_one_line_and_never_raises(tmp_path, monkeypatch, capsys):
    monkeypatch.setattr(exclude_module.sys, "platform", "darwin")

    def broken(*_args, **_kwargs):
        raise OSError("no xattr here")

    monkeypatch.setattr(mark_module.subprocess, "run", broken)
    target = tmp_path / "payloads.sqlite3"
    target.write_text("x")
    exclude_module.exclude_payloads_from_backup(tmp_path)
    err = capsys.readouterr().err
    assert err.count("\n") == 1
    assert "OSError" in err
    assert "no xattr here" not in err
