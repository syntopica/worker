import json

import pytest

from worker.config.load_worker_config import load_worker_config
from worker.store.migrate_state import migrate_state
from worker.store.open_store import open_store

CONFIG = {
    "listen": "127.0.0.1:0",
    "models": {
        "model-a": {"num_ctx": 40960, "keep_alive": "5m", "cold_gb": 30, "warm_gb": 2},
        "model-b": {"num_ctx": 8192, "keep_alive": "5m", "cold_gb": 8, "warm_gb": 1},
    },
    "queues": {
        "pa.bulk": {"run_when": "idle", "max_outstanding": 3},
        "pa.live": {"run_when": "active_ok", "weight": 3},
    },
    "nodes": {
        "node-a": {
            "trust": "owner",
            "memory_budget_gb": 44,
            "ollama_url": "http://127.0.0.1:1",
            "ollama_launchd_label": "label-a",
        },
        "node-g": {
            "trust": "guest",
            "memory_budget_gb": 8,
            "ollama_url": "http://127.0.0.1:1",
            "ollama_launchd_label": "label-g",
        },
    },
    "producers": {"pa": ["pa.bulk", "pa.live"]},
}


@pytest.fixture
def config(tmp_path):
    path = tmp_path / "config.json"
    path.write_text(json.dumps(CONFIG))
    return load_worker_config(path)


def fresh_store(state):
    """Migrate, as the entry points do, then open."""
    migrate_state(state)
    return open_store(state)


@pytest.fixture
def conn(tmp_path):
    return fresh_store(tmp_path / "state")


def body(key="k1", queue="pa.bulk", privacy="mail", **extra):
    data = {
        "contract": 1,
        "kind": "inference",
        "queue": queue,
        "idempotency_key": key,
        "priority": 50,
        "privacy": privacy,
        "requirements": {"capability": "chat.json", "models": ["model-a"]},
        "input": {"messages": [{"role": "user", "content": "hi"}], "options": {"temperature": 0}},
    }
    data.update(extra)
    return json.dumps(data).encode()
