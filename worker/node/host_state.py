"""One sample of the signals that decide whether a machine may work."""

from dataclasses import dataclass


@dataclass(frozen=True)
class HostState:
    """None means the read failed."""

    idle_s: float | None
    on_ac: bool | None
    pressure: str
