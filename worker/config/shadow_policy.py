"""How a queue samples its answers for judged comparison (amendment: shadow sampling)."""

from dataclasses import dataclass


@dataclass(frozen=True)
class ShadowPolicy:
    """``targets`` are ``ollama:<model>``, ``openrouter:<model>`` or ``runner:<profile>``.

    ``rate`` is the share of succeeded inference jobs shadowed, ``judge`` the
    task profile that scores the answers, and ``max_pending`` the unfinished
    shadow members above which no new group opens.
    """

    rate: float
    targets: tuple[str, ...]
    judge: str
    max_pending: int = 20
