"""Build a queue's OpenRouter route from its raw configuration entry."""

from typing import Any

from worker.config.open_router_route import OpenRouterRoute


def parse_openrouter_route(name: str, raw: dict[str, Any] | None) -> OpenRouterRoute | None:
    """None when the queue has no route; raise ValueError on a malformed one.

    Only ``:free`` model ids are accepted: paid endpoints need a budget
    (rung 5), which this route does not carry.
    """
    if raw is None:
        return None
    models = raw.get("models")
    if not isinstance(models, dict) or not models:
        raise ValueError(f"queue {name}: openrouter.models must map local to remote models")
    if not all(isinstance(v, str) and v.endswith(":free") for v in models.values()):
        raise ValueError(f"queue {name}: openrouter models must be :free endpoints")
    fallbacks = raw.get("fallbacks", [])
    if not isinstance(fallbacks, list) or not all(
        isinstance(v, str) and v.endswith(":free") for v in fallbacks
    ):
        raise ValueError(f"queue {name}: openrouter.fallbacks must list :free endpoints")
    cap = raw.get("daily_cap")
    if cap is not None and (isinstance(cap, bool) or not isinstance(cap, int) or cap < 1):
        raise ValueError(f"queue {name}: openrouter.daily_cap must be a positive integer")
    return OpenRouterRoute(
        models=tuple((str(k), str(v)) for k, v in models.items()),
        after_s=float(raw.get("after_s", 0)),
        zdr=bool(raw.get("zdr", True)),
        fallbacks=tuple(fallbacks),
        daily_cap=cap,
    )
