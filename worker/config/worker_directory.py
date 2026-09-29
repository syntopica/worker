"""Where this instance keeps worker's configuration and state."""

import json
from pathlib import Path


def worker_directory(data: Path) -> Path:
    """Return ``worker.path`` from the local file, then the tracked one, else ``worker``.

    A relative value resolves against the file that declared it, the same rule
    the other Syntopica engines follow.
    """
    for name in ("syntopica.local.json", "syntopica.config.json"):
        file = data / name
        if not file.is_file():
            continue
        section = json.loads(file.read_text()).get("worker") or {}
        value = section.get("path")
        if isinstance(value, str) and value:
            return (file.parent / value).resolve()
    return (data / "worker").resolve()
