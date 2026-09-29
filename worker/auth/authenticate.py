"""Resolve a bearer token to a principal."""

import hashlib
import hmac
from pathlib import Path

from worker.auth.load_principals import load_principals
from worker.auth.principal import Principal
from worker.jobs.api_error import ApiError


def authenticate(state_dir: Path, header: str | None) -> Principal:
    """Compare hashes in constant time; raise ApiError(401) when nothing matches."""
    if not header or not header.startswith("Bearer "):
        raise ApiError(401, "unauthorized")
    digest = hashlib.sha256(header.removeprefix("Bearer ").encode()).hexdigest()
    for name, entry in load_principals(state_dir).items():
        if hmac.compare_digest(digest, entry["sha256"]):
            return Principal(entry["kind"], name)
    raise ApiError(401, "unauthorized")
