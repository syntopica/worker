"""Create a principal and its bearer token."""

import hashlib
import json
import re
import secrets
from pathlib import Path

from worker.auth.load_principals import load_principals
from worker.auth.write_private_file import write_private_file

_NAME = re.compile(r"^[a-z0-9][a-z0-9._-]{0,63}$")


def add_principal(state_dir: Path, kind: str, name: str) -> str:
    """Return the plain token; only its SHA-256 is stored in principals.json.

    Raises ValueError for a name that could not safely be a file name.
    """
    if not _NAME.fullmatch(name):
        raise ValueError("invalid principal name")
    token = secrets.token_urlsafe(32)
    principals = load_principals(state_dir)
    principals[name] = {"kind": kind, "sha256": hashlib.sha256(token.encode()).hexdigest()}
    state_dir.mkdir(parents=True, exist_ok=True)
    token_dir = state_dir / "tokens"
    token_dir.mkdir(mode=0o700, exist_ok=True)
    token_dir.chmod(0o700)
    write_private_file(token_dir / f"{name}.token", token, atomic=False)
    write_private_file(state_dir / "principals.json", json.dumps(principals, indent=2), atomic=True)
    return token
