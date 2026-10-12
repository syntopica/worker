import dataclasses
import json
import time

import pytest

from tests.conftest import CONFIG
from tests.node.fake_ollama import FakeOllama
from tests.node.test_run_attempt import PIN, TYPING, Link, lease, nap, states
from worker.config.load_worker_config import load_worker_config
from worker.config.node_policy import NodePolicy
from worker.node import node_step as node_step_module
from worker.node import relieve_owner_active as relieve_module
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory
from worker.node.node_step import node_step
from worker.node.relieve_owner_active import relieve_owner_active
from worker.node.run_attempt import run_attempt


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def _unloads(monkeypatch):
    unloaded = []
    monkeypatch.setattr(
        relieve_module, "unload_model", lambda url, model: unloaded.append(model) or True
    )
    return unloaded


def test_the_flag_is_read_per_node_and_defaults_off(tmp_path):
    raw = json.loads(json.dumps(CONFIG))
    raw["nodes"]["node-a"]["local_idle_only"] = True
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    config = load_worker_config(path)
    assert config.nodes["node-a"].local_idle_only is True
    assert config.nodes["node-g"].local_idle_only is False


def test_owner_activity_unloads_owned_models_once_per_spell(monkeypatch):
    unloaded = _unloads(monkeypatch)
    memory = NodeMemory(owned={"model-a"})
    assert relieve_owner_active("u", memory, ["model-a", "other"], owner_active=True) == ["model-a"]
    assert relieve_owner_active("u", memory, ["model-a"], owner_active=True) == []
    assert unloaded == ["model-a"]
    relieve_owner_active("u", memory, [], owner_active=False)
    assert memory.active_tried == set()


def test_an_idle_owner_or_a_foreign_model_is_left_alone(monkeypatch):
    unloaded = _unloads(monkeypatch)
    assert relieve_owner_active("u", NodeMemory(owned={"m"}), ["m"], owner_active=False) == []
    assert relieve_owner_active("u", NodeMemory(), ["m"], owner_active=True) == []
    assert unloaded == []


def test_active_ok_work_is_released_on_input_on_a_local_idle_only_node(fake):
    fake.delay = 3.0
    link = Link()
    node = NodePolicy("node-a", "owner", 300, 44, fake.url, "label-a", local_idle_only=True)
    outcome = run_attempt(lease("active_ok"), link, PIN, node, states(TYPING), nap, time.time)
    assert outcome == "preempted"
    assert link.completed[0]["error_code"] == "user_active"


def test_a_local_idle_only_node_neither_leases_nor_keeps_its_model_while_active(
    config, monkeypatch
):
    unloaded = _unloads(monkeypatch)
    monkeypatch.setattr(node_step_module, "resident_models", lambda url: ["model-a"])
    nodes = dict(config.nodes)
    nodes["node-a"] = dataclasses.replace(nodes["node-a"], local_idle_only=True)
    config = dataclasses.replace(config, nodes=nodes)

    class L:
        def report(self, state):
            pass

        def lease(self, *args):
            raise AssertionError("must not lease the local model while the owner is active")

    memory = NodeMemory(owned={"model-a"})
    rest = node_step(
        config, "node-a", L(), memory, lambda: HostState(30, True, "normal"), nap, lambda: 1000.0
    )
    assert rest is True
    assert unloaded == ["model-a"]
