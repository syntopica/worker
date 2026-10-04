"""The validated filters of one admin job listing request."""

from dataclasses import dataclass


@dataclass(frozen=True)
class JobsFilter:
    """``None`` leaves a field unfiltered; ``limit`` is already clamped to 1..100."""

    queue: str | None
    state: str | None
    producer: str | None
    before: str | None
    limit: int
