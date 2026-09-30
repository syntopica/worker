"""Whether a manifest entry stays inside the root it is relative to."""

from pathlib import PurePosixPath


def is_safe_relative_path(value: object) -> bool:
    """A non-empty relative POSIX path with no ``..`` and no empty or ``.`` parts."""
    if not isinstance(value, str) or not value or "\\" in value or "\x00" in value:
        return False
    path = PurePosixPath(value)
    if path.is_absolute():
        return False
    return all(part not in {"", ".", ".."} for part in value.split("/"))
