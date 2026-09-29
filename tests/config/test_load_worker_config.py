import json

import pytest

from worker.config.load_worker_config import load_worker_config

CONFIG = {
    "listen": "127.0.0.1:8765",
    "models": {"model-a": {"num_ctx": 40960, "keep_alive": "5m", "cold_gb": 30, "warm_gb": 2}},
    "queues": {
        "producer-a.bulk": {"run_when": "idle"},
        "producer-a.live": {"run_when": "active_ok", "weight": 3},
    },
    "nodes": {
        "node-a": {
            "trust": "owner",
            "memory_budget_gb": 44,
            "ollama_url": "http://127.0.0.1:11434",
            "ollama_launchd_label": "label-a",
        }
    },
    "producers": {"producer-a": ["producer-a.bulk", "producer-a.live"]},
}


def write(tmp_path, data):
    path = tmp_path / "config.json"
    path.write_text(json.dumps(data))
    return path


def test_defaults_fill_every_optional_value(tmp_path):
    config = load_worker_config(write(tmp_path, CONFIG))
    assert (config.listen_host, config.listen_port) == ("127.0.0.1", 8765)
    bulk = config.queues["producer-a.bulk"]
    assert (bulk.weight, bulk.max_outstanding, bulk.unacked_ttl_hours) == (1, 200, 72)
    assert config.nodes["node-a"].idle_threshold_s == 300
    assert config.max_payload_bytes == 1_048_576
    assert (config.max_split, config.max_parked_runs, config.split_after_preemptions) == (8, 3, 3)


def test_sensitive_classes_default_to_local_executors_on_owner_nodes(tmp_path):
    config = load_worker_config(write(tmp_path, CONFIG))
    for sensitive in ("personal", "mail", "secret"):
        assert config.privacy[sensitive] == frozenset({"ollama", "local-cpu"})
        assert config.trust[sensitive] == frozenset({"owner"})


def test_an_override_replaces_one_class_only(tmp_path):
    data = {**CONFIG, "privacy": {"mail": {"executors": ["ollama"], "trust": ["owner"]}}}
    config = load_worker_config(write(tmp_path, data))
    assert config.privacy["mail"] == frozenset({"ollama"})
    assert config.privacy["secret"] == frozenset({"ollama", "local-cpu"})


def test_unknown_run_when_is_refused(tmp_path):
    data = {**CONFIG, "queues": {"q": {"run_when": "sometimes"}}}
    with pytest.raises(ValueError, match="run_when"):
        load_worker_config(write(tmp_path, data))
