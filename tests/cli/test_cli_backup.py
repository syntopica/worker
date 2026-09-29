import sqlite3
import stat

from tests.cli.test_cli import instance
from tests.conftest import body
from worker.cli.main import main
from worker.jobs.submit_job import submit_job
from worker.store.open_store import open_store


def rows(path):
    with sqlite3.connect(path) as conn:
        return conn.execute("SELECT count(*) FROM jobs").fetchone()[0]


def test_a_row_written_before_the_backup_survives(tmp_path, monkeypatch, config):
    state = instance(tmp_path, monkeypatch)
    conn = open_store(state)
    submit_job(conn, config, "pa", body(), 1.0)
    conn.close()
    dest = tmp_path / "b.sqlite3"
    assert main(["backup", str(dest)]) == 0
    assert rows(dest) == 1


def test_uncheckpointed_wal_rows_are_backed_up_with_the_source_open(tmp_path, monkeypatch, config):
    state = instance(tmp_path, monkeypatch)
    conn = open_store(state)
    conn.execute("PRAGMA wal_autocheckpoint=0")
    for n in range(3):
        submit_job(conn, config, "pa", body(key=f"k{n}"), 1.0)
    assert (state / "meta.sqlite3-wal").stat().st_size > 0
    dest = tmp_path / "b.sqlite3"
    assert main(["backup", str(dest)]) == 0
    assert rows(dest) == 3
    conn.close()


def test_absent_store_is_exit_2_and_creates_nothing(tmp_path, monkeypatch, capsys):
    state = instance(tmp_path, monkeypatch)
    dest = tmp_path / "b.sqlite3"
    assert main(["backup", str(dest)]) == 2
    assert capsys.readouterr().err == f"worker: no store at {state / 'meta.sqlite3'}\n"
    assert not dest.exists()
    assert not state.exists()


def test_backup_file_mode_is_0600(tmp_path, monkeypatch):
    state = instance(tmp_path, monkeypatch)
    open_store(state).close()
    dest = tmp_path / "b.sqlite3"
    assert main(["backup", str(dest)]) == 0
    assert stat.S_IMODE(dest.stat().st_mode) == 0o600


def test_existing_dest_is_refused_and_untouched(tmp_path, monkeypatch, capsys):
    state = instance(tmp_path, monkeypatch)
    open_store(state).close()
    dest = tmp_path / "b.sqlite3"
    dest.write_text("keep me")
    assert main(["backup", str(dest)]) == 2
    assert "already exists" in capsys.readouterr().err
    assert dest.read_text() == "keep me"
