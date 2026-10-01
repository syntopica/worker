"""Remove credentials from text bound for a remote executor (owner rule 2026-10-01)."""

from worker.node.credential_patterns import CREDENTIAL_PATTERNS
from worker.node.mask_secret import mask_secret


def redact_credentials(text: str) -> str:
    """Return ``text`` with every credential pattern's secret replaced by a fixed mark."""
    for pattern in CREDENTIAL_PATTERNS:
        text = pattern.sub(mask_secret, text)
    return text
