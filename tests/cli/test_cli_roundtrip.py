import json
import threading

import pytest

from tests.conftest import CONFIG, body
from worker.api.build_server import build_server
from worker.cli.main import main


@pytest.fixture
def served(config, tmp_path, monkeypatch):
    state = tmp_path / "worker" / "state"
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


def run(capsys, *argv):
    code = main(list(argv))
    captured = capsys.readouterr()
    return code, captured.out, captured.err


def test_submit_jobs_cancel_and_status_round_trip(served, capsys):
    for kind, name in (("producer", "pa"), ("admin", "admin")):
        assert run(capsys, "token", "add", "--kind", kind, "--name", name)[0] == 0
    job_file = served / "job.json"
    job_file.write_bytes(body())

    code, out, _ = run(capsys, "submit", "--producer", "pa", str(job_file))
    submitted = json.loads(out)
    assert (code, submitted["created"]) == (0, True)

    code, out, _ = run(capsys, "jobs", "--producer", "pa", "--queue", "pa.bulk")
    assert (code, json.loads(out)) == (0, [])

    code, out, _ = run(capsys, "cancel", "--producer", "pa", submitted["id"])
    assert (code, json.loads(out)) == (0, {"state": "cancelled"})

    code, out, _ = run(capsys, "status", "--json")
    status = json.loads(out)
    assert code == 0
    assert status["queues"]["pa.bulk"]["states"] == {"cancelled": 1}


def test_a_refusal_is_one_stderr_line_and_exit_1(served, capsys):
    run(capsys, "token", "add", "--kind", "producer", "--name", "pa")
    assert run(capsys, "token", "add", "--kind", "producer", "--name", "admin")[0] == 0
    code, out, err = run(capsys, "status")
    assert (code, out, err) == (1, "", "worker: 403 forbidden\n")


def test_cancelling_an_unknown_job_reports_the_code(served, capsys):
    run(capsys, "token", "add", "--kind", "producer", "--name", "pa")
    code, _, err = run(capsys, "cancel", "--producer", "pa", "nope")
    assert code == 1
    assert err.startswith("worker: 404 ")
