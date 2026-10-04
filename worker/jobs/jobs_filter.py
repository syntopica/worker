"""The validated filters of one admin job listing request."""

from dataclasses import dataclass


@dataclass(frozen=True)
class JobsFilter:
    """``None`` leaves a field unfiltered; a job matches any one of ``states``.

    ``limit`` is already clamped to 1..100.
    """

    queue: str | None
    states: tuple[str, ...] | None
    producer: str | None
    before: str | None
    limit: int
