"""Create a principal and its bearer token."""

import hashlib
import json
import secrets
from pathlib import Path

from worker.auth.load_principals import load_principals


def add_principal(state_dir: Path, kind: str, name: str) -> str:
    """Return the plain token; only its SHA-256 is stored in principals.json."""
    token = secrets.token_urlsafe(32)
    principals = load_principals(state_dir)
    principals[name] = {"kind": kind, "sha256": hashlib.sha256(token.encode()).hexdigest()}
    state_dir.mkdir(parents=True, exist_ok=True)
    path = state_dir / "principals.json"
    path.write_text(json.dumps(principals, indent=2))
    path.chmod(0o600)
    token_dir = state_dir / "tokens"
    token_dir.mkdir(mode=0o700, exist_ok=True)
    token_file = token_dir / f"{name}.token"
    token_file.write_text(token)
    token_file.chmod(0o600)
    return token
