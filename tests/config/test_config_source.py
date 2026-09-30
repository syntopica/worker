import json
import os

from tests.conftest import CONFIG
from worker.config.config_source import ConfigSource
from worker.config.load_worker_config import load_worker_config


def write(path, data, stamp):
    path.write_text(json.dumps(data))
    os.utime(path, ns=(stamp, stamp))


def test_a_changed_file_is_reloaded_and_an_unchanged_one_is_not(tmp_path):
    path = tmp_path / "config.json"
    write(path, CONFIG, 1_000_000_000)
    source = ConfigSource(path, load_worker_config(path))
    first = source.current()
    assert source.current() is first
    grown = {**CONFIG, "queues": {**CONFIG["queues"], "pa.new": {"run_when": "idle"}}}
    write(path, grown, 2_000_000_000)
    assert "pa.new" in source.current().queues


def test_an_invalid_new_file_keeps_the_last_good_config(tmp_path, capsys):
    path = tmp_path / "config.json"
    write(path, CONFIG, 1_000_000_000)
    source = ConfigSource(path, load_worker_config(path))
    path.write_text("{not json")
    os.utime(path, ns=(2_000_000_000, 2_000_000_000))
    assert set(source.current().queues) == set(CONFIG["queues"])
    assert capsys.readouterr().err == "worker: config reload refused: JSONDecodeError\n"
