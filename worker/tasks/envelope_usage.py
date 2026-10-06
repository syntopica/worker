"""Token counts from the ``usage`` object of a runner's JSON envelope."""

import json
from typing import Any


def envelope_usage(stdout: str, in_key: str, out_key: str) -> dict[str, Any]:
    """``tokens_in`` and ``tokens_out`` where the envelope carries them as integers.

    A malformed envelope, a missing ``usage`` or a non-integer count is left out
    rather than guessed, so the ledger stores null instead of a wrong number.
    No runner envelope carries a price, so ``cost_usd`` is never set here.
    """
    try:
        envelope = json.loads(stdout)
    except ValueError:
        return {}
    usage = envelope.get("usage") if isinstance(envelope, dict) else None
    if not isinstance(usage, dict):
        return {}
    counts = {"tokens_in": usage.get(in_key), "tokens_out": usage.get(out_key)}
    return {k: v for k, v in counts.items() if isinstance(v, int) and not isinstance(v, bool)}
