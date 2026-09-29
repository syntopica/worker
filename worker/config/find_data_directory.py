"""Walk upward from a directory to the syntopica instance that contains it."""

from pathlib import Path

CONFIG_NAME = "syntopica.config.json"


def find_data_directory(start: Path) -> Path | None:
    """Return the nearest ancestor holding ``syntopica.config.json``, or None.

    The walk stops at a repository boundary: a directory carrying ``.git`` that
    is not itself an instance ends the search, because a command run inside one
    repository must not silently read the instance that happens to enclose it.
    The instance is a repository too, so the config check comes first.
    """
    current = start.resolve()
    for directory in (current, *current.parents):
        if (directory / CONFIG_NAME).is_file():
            return directory
        if (directory / ".git").exists():
            return None
    return None
