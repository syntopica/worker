"""A stable hash of a submit body, for idempotency comparisons."""

import hashlib
import json
from typing import Any


def payload_hash(body: dict[str, Any]) -> str:
    """SHA-256 of the canonical JSON (sorted keys, no whitespace)."""
    canonical = json.dumps(body, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode()).hexdigest()
