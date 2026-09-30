import json
from pathlib import Path

from worker.config.load_worker_config import load_worker_config
from worker.node.effective_pressure import effective_pressure
from worker.node.parse_free_pct import parse_free_pct
from worker.node.sample_host_state import sample_host_state


def fake_host(level: str | None, free: str | None):
    outputs = {
        "kern.memorystatus_vm_pressure_level": level,
        "kern.memorystatus_level": free,
        "IOHIDSystem": '"HIDIdleTime" = 900000000000',
        "ps": "Now drawing from 'AC Power'",
    }

    def run(args: list[str]) -> str | None:
        return next(v for k, v in outputs.items() if k in args)

    return run


def test_a_kernel_warning_with_a_third_free_is_not_pressure():
    # Measured 2026-09-30: level 2 with 28-30% free while the node's own model
    # was the largest resident; acting on it unloaded that model in a loop.
    assert sample_host_state(fake_host("2", "30"), min_free_pct=15).pressure == "normal"


def test_a_kernel_warning_counts_once_free_memory_is_low():
    assert sample_host_state(fake_host("2", "12"), min_free_pct=15).pressure == "warn"


def test_critical_always_counts():
    assert effective_pressure("critical", 80.0, 15.0) == "critical"
    assert effective_pressure("critical", None, 15.0) == "critical"


def test_an_unreadable_signal_still_blocks_work():
    assert sample_host_state(fake_host("2", None)).pressure == "unknown"
    assert sample_host_state(fake_host("2", "n/a")).pressure == "unknown"
    assert sample_host_state(fake_host(None, "50")).pressure == "unknown"
    assert effective_pressure("unknown", 50.0, 15.0) == "unknown"


def test_normal_stays_normal_whatever_is_free():
    assert effective_pressure("normal", 5.0, 15.0) == "normal"


def test_parse_free_pct_rejects_what_is_not_a_percentage():
    assert parse_free_pct("30\n") == 30.0
    assert parse_free_pct("101") is None
    assert parse_free_pct("-1") is None


def test_min_free_pct_is_read_per_node_with_a_default(tmp_path: Path):
    node = {
        "trust": "owner",
        "memory_budget_gb": 44,
        "ollama_url": "http://127.0.0.1:11434",
        "ollama_launchd_label": "label",
    }
    path = tmp_path / "config.json"
    raw = {"models": {}, "queues": {}, "producers": {}}
    raw["nodes"] = {"a": node, "b": {**node, "min_free_pct": 8}}
    path.write_text(json.dumps(raw))
    config = load_worker_config(path)
    assert config.nodes["a"].min_free_pct == 15.0
    assert config.nodes["b"].min_free_pct == 8.0
