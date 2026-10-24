"""Token counts from an agy final result, which carries the cumulative usage."""

import json
from typing import Any

from worker.tasks.agy_final_result import agy_final_result
from worker.tasks.envelope_usage import envelope_usage


def agy_usage(stdout: str) -> dict[str, Any]:
    """Only the final result's ``usage`` counts: step updates are never summed."""
    final = agy_final_result(stdout)
    if final is None:
        return {}
    return envelope_usage(json.dumps(final[0]), "input_tokens", "output_tokens")
