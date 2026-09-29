import json
import sqlite3

from tests.conftest import CONFIG, fresh_store
from worker.cli.main import main


def instance(tmp_path, monkeypatch):
    (tmp_path / "syntopica.config.json").write_text("{}")
    (tmp_path / "worker").mkdir()
    (tmp_path / "worker" / "config.json").write_text(json.dumps(CONFIG))
    monkeypatch.setenv("SYNTOPICA_DATA", str(tmp_path))
    return tmp_path / "worker" / "state"


def test_token_add_prints_once_and_stores_only_a_hash(tmp_path, monkeypatch, capsys):
    state = instance(tmp_path, monkeypatch)
    assert main(["token", "add", "--kind", "admin", "--name", "admin"]) == 0
    token = capsys.readouterr().out.strip()
    assert token not in (state / "principals.json").read_text()


def test_backup_copies_metadata_only(tmp_path, monkeypatch):
    state = instance(tmp_path, monkeypatch)
    fresh_store(state).close()
    dest = tmp_path / "backup.sqlite3"
    assert main(["backup", str(dest)]) == 0
    tables = {
        r[0]
        for r in sqlite3.connect(dest).execute("SELECT name FROM sqlite_master WHERE type='table'")
    }
    assert "jobs" in tables
    assert "inputs" not in tables


def test_token_add_rejects_a_traversal_name_with_exit_2(tmp_path, monkeypatch, capsys):
    state = instance(tmp_path, monkeypatch)
    assert main(["token", "add", "--kind", "producer", "--name", "../x"]) == 2
    assert "invalid principal name" in capsys.readouterr().err
    assert not (state / "tokens").exists() or not list((state / "tokens").iterdir())


def test_missing_token_is_exit_2_without_a_traceback(tmp_path, monkeypatch, capsys):
    instance(tmp_path, monkeypatch)
    assert main(["status"]) == 2
    assert capsys.readouterr().err.startswith("worker: no token for 'admin'")
