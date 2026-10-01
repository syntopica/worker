"""How a queue escalates inference to OpenRouter's free endpoints (spec 8, rung 3)."""

from dataclasses import dataclass


@dataclass(frozen=True)
class OpenRouterRoute:
    """``models`` maps a local model to the OpenRouter model that may stand in for it.

    A job escalates only after waiting ``after_s`` since it was created. With
    ``zdr`` (the default) a non-public job is sent only to zero-data-retention
    endpoints; turning it off is an explicit owner decision for the queue.
    ``fallbacks`` are further free models OpenRouter tries, in order, when
    the mapped one errors or is rate limited.
    """

    models: tuple[tuple[str, str], ...]
    after_s: float
    zdr: bool = True
    fallbacks: tuple[str, ...] = ()
