import json
import threading

import pytest

from tests.conftest import CONFIG
from worker.api.build_server import build_server
from worker.cli.main import main
from worker.store.migrate_state import migrate_state


@pytest.fixture
def served(config, tmp_path, monkeypatch):
    state = tmp_path / "worker" / "state"
    migrate_state(state)
    server = build_server(config, state)  # port 0: the OS picks a free one
    threading.Thread(target=server.serve_forever, daemon=True).start()
    (tmp_path / "worker").mkdir(exist_ok=True)
    (tmp_path / "syntopica.config.json").write_text("{}")
    port = server.server_address[1]
    cfg = {**CONFIG, "listen": f"127.0.0.1:{port}"}
    (tmp_path / "worker" / "config.json").write_text(json.dumps(cfg))
    monkeypatch.setenv("SYNTOPICA_DATA", str(tmp_path))
    yield tmp_path
    server.shutdown()
    server.server_close()


@pytest.fixture
def run(capsys):
    def invoke(*argv):
        code = main(list(argv))
        captured = capsys.readouterr()
        return code, captured.out, captured.err

    return invoke
