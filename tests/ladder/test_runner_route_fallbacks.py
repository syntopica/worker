import json

import pytest

from tests.conftest import CONFIG, body, fresh_store
from tests.ladder.test_model_cooldown import lease
from worker.config.load_worker_config import load_worker_config
from worker.jobs.submit_job import submit_job


def make_config(tmp_path, route):
    raw = json.loads(json.dumps(CONFIG))
    raw["profiles"] = {
        "a.gemini": {"runner": "agy", "model": "gemini-pro", "privacy": ["mail"]},
        "a.claude": {"runner": "agy", "model": "claude-sonnet", "privacy": ["mail"]},
    }
    raw["queues"]["pa.bulk"]["runner"] = route
    raw["privacy"] = {"mail": {"executors": ["ollama", "runner"], "trust": ["owner"]}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def test_a_resting_route_profile_hands_the_job_to_its_fallback(tmp_path):
    config = make_config(tmp_path, {"profile": "a.gemini", "fallbacks": ["a.claude"]})
    conn = fresh_store(tmp_path / "state")
    conn.execute("INSERT INTO cooldowns (runner, until) VALUES ('agy:gemini-pro', 100.0)")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, 1.0).model == "a.claude"


def test_the_route_profile_comes_first_while_it_can_run(tmp_path):
    config = make_config(tmp_path, {"profile": "a.gemini", "fallbacks": ["a.claude"]})
    conn = fresh_store(tmp_path / "state")
    submit_job(conn, config, "pa", body(), 0.0)
    assert lease(conn, config, 1.0).model == "a.gemini"


@pytest.mark.parametrize("fallbacks", ["a.claude", ["missing"], [""]])
def test_a_malformed_or_unknown_route_fallback_is_refused(tmp_path, fallbacks):
    with pytest.raises(ValueError, match="runner"):
        make_config(tmp_path, {"profile": "a.gemini", "fallbacks": fallbacks})
