"""An authorized way to run one runner over one input root (spec 8)."""

from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class TaskProfile:
    """``nodes`` None means any node; ``input_root`` None means no inputs.

    ``command`` is the runner executable (default: its name on PATH) and
    ``env_unset`` the variables removed from its environment, such as an API
    key that would bill a call the subscription login should carry.
    """

    name: str
    runner: str
    model: str | None
    reasoning: str | None
    privacy: frozenset[str]
    timeout_s: float
    nodes: frozenset[str] | None
    input_root: Path | None
    command: str | None = None
    env_unset: frozenset[str] = frozenset()
