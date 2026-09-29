"""Who is calling: a producer, a node or the admin CLI."""

from dataclasses import dataclass


@dataclass(frozen=True)
class Principal:
    """``kind`` is ``producer``, ``node`` or ``admin``."""

    kind: str
    name: str
