"""A queued job as the scheduler sees it: no payload, only what ranking needs."""

from dataclasses import dataclass


@dataclass(frozen=True)
class Candidate:
    """``parked_min_idle_s`` is the current idle a parked job needs before it runs."""

    job_id: str
    queue: str
    model: str
    priority: int
    privacy: str
    created: float
    parked: bool
    parked_min_idle_s: float
