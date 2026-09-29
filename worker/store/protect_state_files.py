"""Owner-only permissions on the state directory and its database files."""

from pathlib import Path

_FILES = ("meta.sqlite3", "payloads.sqlite3")
_SUFFIXES = ("", "-wal", "-shm", "-journal")


def protect_state_files(state_dir: Path) -> None:
    """Directory 0700; each database file and its sidecars 0600, when present."""
    state_dir.chmod(0o700)
    for name in _FILES:
        for suffix in _SUFFIXES:
            path = state_dir / (name + suffix)
            if path.exists():
                path.chmod(0o600)
