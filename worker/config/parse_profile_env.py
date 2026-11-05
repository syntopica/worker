"""Read a task profile's ``env`` object into the pairs a runner is started with."""

from pathlib import Path
from typing import Any


def parse_profile_env(name: str, raw: Any) -> tuple[tuple[str, str], ...]:
    """Sorted string pairs; a leading ``~/`` expands to the home directory.

    Raise ValueError unless ``raw`` is absent or an object of strings.
    """
    if raw is None:
        return ()
    if not isinstance(raw, dict) or not all(
        isinstance(k, str) and k and isinstance(v, str) for k, v in raw.items()
    ):
        raise ValueError(f"profile {name}: env must map variable names to strings")
    return tuple(
        sorted((k, str(Path.home() / v[2:]) if v.startswith("~/") else v) for k, v in raw.items())
    )
