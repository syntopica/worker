"""How a queue sends its inference jobs to a runner first (executor ladder)."""

from dataclasses import dataclass


@dataclass(frozen=True)
class RunnerRoute:
    """``profile`` is the task profile that runs the job; ``after_s`` its wait since creation."""

    profile: str
    after_s: float = 0.0
