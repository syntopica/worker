"""Read the token hashes kept in the untracked state directory."""

import json
from pathlib import Path


def load_principals(state_dir: Path) -> dict[str, dict[str, str]]:
    """``{name: {"kind": ..., "sha256": ...}}``; empty when none exist yet."""
    path = state_dir / "principals.json"
    return json.loads(path.read_text()) if path.is_file() else {}
