import json

from worker.config.worker_directory import worker_directory


def test_local_file_outranks_tracked_config(tmp_path):
    (tmp_path / "syntopica.config.json").write_text(json.dumps({"worker": {"path": "w"}}))
    (tmp_path / "syntopica.local.json").write_text(json.dumps({"worker": {"path": "/abs/w"}}))
    assert str(worker_directory(tmp_path)) == "/abs/w"


def test_default_is_worker_inside_the_instance(tmp_path):
    (tmp_path / "syntopica.config.json").write_text("{}")
    assert worker_directory(tmp_path) == (tmp_path / "worker").resolve()
