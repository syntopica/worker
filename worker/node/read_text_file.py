"""Read a small kernel file without raising."""

from pathlib import Path


def read_text_file(path: str) -> str | None:
    """The file's text, or None when it cannot be read."""
    try:
        return Path(path).read_text()
    except OSError:
        return None
