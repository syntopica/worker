import json

from tests.conftest import CONFIG, fresh_store
from tests.jobs.test_task_fallbacks import lease, task
from worker.config.load_worker_config import load_worker_config
from worker.jobs.submit_job import submit_job


def make_config(tmp_path, max_concurrent):
    raw = json.loads(json.dumps(CONFIG))
    raw["queues"]["pa.syn"] = {"run_when": "idle", "profiles": ["pa.cursor"]}
    raw["producers"]["pa"].append("pa.syn")
    raw["profiles"] = {"pa.cursor": {"runner": "cursor", "privacy": ["internal"]}}
    raw["runners"] = {"cursor": {"max_concurrent": max_concurrent}}
    path = tmp_path / "config.json"
    path.write_text(json.dumps(raw))
    return load_worker_config(path)


def test_a_runner_at_its_concurrency_cap_takes_no_further_task(tmp_path):
    config = make_config(tmp_path, 1)
    conn = fresh_store(tmp_path / "state")
    for key in ("t1", "t2"):
        submit_job(conn, config, "pa", task(key), 0.0)
    assert lease(conn, config, 1.0) is not None
    assert lease(conn, config, 2.0) is None


def test_without_a_cap_a_runner_takes_every_ready_task(tmp_path):
    config = make_config(tmp_path, 2)
    conn = fresh_store(tmp_path / "state")
    for key in ("t1", "t2"):
        submit_job(conn, config, "pa", task(key), 0.0)
    assert lease(conn, config, 1.0) is not None
    assert lease(conn, config, 2.0) is not None
