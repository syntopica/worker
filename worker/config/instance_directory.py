"""The syntopica data directory this invocation serves, if there is one."""

import os
from collections.abc import Mapping
from pathlib import Path

from worker.config.find_data_directory import find_data_directory


def instance_directory(
    environ: Mapping[str, str] | None = None, cwd: Path | None = None
) -> Path | None:
    """``SYNTOPICA_DATA`` first, then the instance enclosing ``cwd``, else None."""
    env = os.environ if environ is None else environ
    data = env.get("SYNTOPICA_DATA")
    if data:
        return Path(data).expanduser().resolve()
    return find_data_directory(Path.cwd() if cwd is None else cwd)
