"""Locate config.json and the state directory for this invocation."""

from pathlib import Path

from worker.config.instance_directory import instance_directory
from worker.config.worker_directory import worker_directory


def resolve_paths() -> tuple[Path, Path]:
    """Raise SystemExit when no Syntopica instance can be found."""
    data = instance_directory()
    if data is None:
        raise SystemExit("worker: no Syntopica instance (set SYNTOPICA_DATA)")
    base = worker_directory(data)
    return base / "config.json", base / "state"
