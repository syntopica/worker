"""What a node reports when an attempt ends."""

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class CompletionReport:
    """``error_code`` is allowlisted: never a provider body or generated text."""

    outcome: str
    output: dict[str, Any] | None
    usage: dict[str, Any]
    executor: dict[str, Any]
    error_code: str | None
    wall_s: float
