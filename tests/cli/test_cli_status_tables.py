import json

import pytest

from tests.cli.test_cli_roundtrip import run, served  # noqa: F401
from tests.conftest import body
from worker.cli.read_token import read_token
from worker.node.coordinator_link import CoordinatorLink


@pytest.fixture
def seeded(served, capsys):  # noqa: F811
    state = served / "worker" / "state"
    for kind, name in (("producer", "pa"), ("admin", "admin"), ("node", "node-a")):
        run(capsys, "token", "add", "--kind", kind, "--name", name)
    (served / "job.json").write_bytes(body())
    run(capsys, "submit", "--producer", "pa", str(served / "job.json"))
    listen = json.loads((served / "worker" / "config.json").read_text())["listen"]
    CoordinatorLink(f"http://{listen}", read_token(state, "node-a"), "node-a").report(
        {
            "reason": "user_active",
            "idle_s": 3.0,
            "on_ac": True,
            "pressure": "normal",
            "resident": ["model-a"],
            "unexpected": [],
        }
    )
    return served


def test_status_table_lists_each_queue(seeded, capsys):
    code, out, _ = run(capsys, "status")
    assert code == 0
    assert "pa.bulk" in out
    assert "done_1h=0" in out


def test_nodes_table_lists_each_reporting_node(seeded, capsys):
    code, out, _ = run(capsys, "nodes")
    assert code == 0
    assert "node-a" in out
    assert "user_active" in out
    assert "resident=['model-a']" in out
