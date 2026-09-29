from tests.conftest import fresh_store
from worker.store.transaction import transaction


def test_two_files_with_the_required_pragmas(tmp_path):
    conn = fresh_store(tmp_path)
    assert conn.execute("PRAGMA journal_mode").fetchone()[0] == "wal"
    assert conn.execute("PRAGMA synchronous").fetchone()[0] == 2  # FULL
    assert conn.execute("PRAGMA p.secure_delete").fetchone()[0] == 1
    assert (tmp_path / "meta.sqlite3").exists()
    assert (tmp_path / "payloads.sqlite3").exists()
    tables = {r[0] for r in conn.execute("SELECT name FROM p.sqlite_master WHERE type='table'")}
    assert tables == {"inputs", "outputs"}


def test_main_file_holds_no_payload_table(tmp_path):
    conn = fresh_store(tmp_path)
    main = {r[0] for r in conn.execute("SELECT name FROM main.sqlite_master WHERE type='table'")}
    assert "inputs" not in main
    assert "outputs" not in main


def test_transaction_rolls_back_on_error(tmp_path):
    conn = fresh_store(tmp_path)
    try:
        with transaction(conn):
            conn.execute("INSERT INTO nodes (name, report, updated) VALUES ('n', '{}', 0)")
            raise RuntimeError("boom")
    except RuntimeError:
        pass
    assert conn.execute("SELECT count(*) FROM nodes").fetchone()[0] == 0
