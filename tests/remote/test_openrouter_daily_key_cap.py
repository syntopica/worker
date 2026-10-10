import json

import pytest

from tests.conftest import CONFIG
from worker.config.load_worker_config import load_worker_config
from worker.policy.capped_queues import capped_queues


def make_config(tmp_path, key_cap):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.bulk"]["openrouter"] = {
        "models": {"model-a": "vendor/model:free"},
        "daily_key_cap": key_cap,
    }
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def test_the_key_cap_counts_every_queue_spend(tmp_path):
    config = make_config(tmp_path, 850)
    assert "pa.bulk" in capped_queues(config, {"pa.bulk": 300, "other.queue": 550})


def test_below_the_key_cap_the_queue_keeps_its_rung(tmp_path):
    config = make_config(tmp_path, 850)
    assert "pa.bulk" not in capped_queues(config, {"pa.bulk": 800, "other.queue": 49})


def test_a_key_cap_must_be_a_positive_integer(tmp_path):
    with pytest.raises(ValueError, match="daily_key_cap"):
        make_config(tmp_path, 0)
