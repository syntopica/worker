"""The live configuration: reloaded when ``config.json`` changes on disk."""

import sys
from pathlib import Path

from worker.config.load_worker_config import load_worker_config
from worker.config.worker_config import WorkerConfig


class ConfigSource:
    """``current()`` stats the file and reloads it when its mtime or size moved.

    An invalid new file is refused with one line on stderr (the exception
    class only) and the last good configuration stays in force. The listen
    address is read once at startup: changing it still needs a restart.
    """

    def __init__(self, path: Path, initial: WorkerConfig) -> None:
        self.path = path
        self._config = initial
        self._stamp = self._read_stamp()

    def _read_stamp(self) -> tuple[int, int] | None:
        try:
            info = self.path.stat()
        except OSError:
            return None
        return (info.st_mtime_ns, info.st_size)

    def current(self) -> WorkerConfig:
        """The configuration in force now."""
        stamp = self._read_stamp()
        if stamp is None or stamp == self._stamp:
            return self._config
        self._stamp = stamp
        try:
            self._config = load_worker_config(self.path)
        except (OSError, ValueError, KeyError, TypeError) as error:
            print(f"worker: config reload refused: {type(error).__name__}", file=sys.stderr)
        return self._config
