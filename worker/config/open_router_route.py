"""How a queue escalates inference to OpenRouter's free endpoints (spec 8, rung 3)."""

from dataclasses import dataclass


@dataclass(frozen=True)
class OpenRouterRoute:
    """``models`` maps a local model to the OpenRouter model that may stand in for it.

    A job escalates only after waiting ``after_s`` since it was created. With
    ``zdr`` (the default) a non-public job is sent only to zero-data-retention
    endpoints; turning it off is an explicit owner decision for the queue.
    ``fallbacks`` are further free models OpenRouter tries, in order, when
    the mapped one errors or is rate limited. ``daily_cap`` bounds the
    queue's OpenRouter attempts per UTC day; None leaves it unbounded.
    ``daily_key_cap`` bounds the key instead: the queue stops once every
    queue's OpenRouter attempts that UTC day reach it, so a bulk queue can use
    what the others leave of the key's allowance and still leave them a reserve.
    """

    models: tuple[tuple[str, str], ...]
    after_s: float
    zdr: bool = True
    fallbacks: tuple[str, ...] = ()
    daily_cap: int | None = None
    daily_key_cap: int | None = None
