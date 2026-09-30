import json

import pytest

from tests.conftest import body
from worker.cli.read_token import read_token
from worker.node.coordinator_link import CoordinatorLink


@pytest.fixture
def seeded(served, run):
    state = served / "worker" / "state"
    for kind, name in (("producer", "pa"), ("admin", "admin"), ("node", "node-a")):
        run("token", "add", "--kind", kind, "--name", name)
    (served / "job.json").write_bytes(body())
    run("submit", "--producer", "pa", str(served / "job.json"))
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


def test_status_table_lists_each_queue(seeded, run):
    code, out, _ = run("status")
    assert code == 0
    assert "pa.bulk" in out
    assert "done_1h=0" in out


def test_nodes_table_lists_each_reporting_node(seeded, run):
    code, out, _ = run("nodes")
    assert code == 0
    assert "node-a" in out
    assert "user_active" in out
    assert "resident=['model-a']" in out


def test_quality_and_costs_print_once_there_are_attempts(seeded, run):
    assert run("quality")[0] == 0
    code, out, _ = run("quality", "--json")
    assert code == 0
    assert json.loads(out) == {"attempts": [], "ratings": []}
    assert run("costs", "--json")[0] == 0
