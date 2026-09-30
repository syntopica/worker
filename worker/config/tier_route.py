"""What one quality tier of a queue runs on (spec amendment: quality tiers)."""

from dataclasses import dataclass


@dataclass(frozen=True)
class TierRoute:
    """``models`` are tried before the producer's list; ``profiles`` maps a granted profile to its stand-in."""

    models: tuple[str, ...] = ()
    profiles: tuple[tuple[str, str], ...] = ()
