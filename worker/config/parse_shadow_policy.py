"""Build a queue's shadow policy from its raw configuration entry."""

from typing import Any

from worker.config.shadow_policy import ShadowPolicy

_KINDS = ("ollama", "openrouter", "runner")


def parse_shadow_policy(name: str, raw: dict[str, Any] | None) -> ShadowPolicy | None:
    """None when the queue has none; raise ValueError on a malformed one.

    OpenRouter targets must be ``:free`` endpoints, as in a route.
    """
    if raw is None:
        return None
    rate = float(raw.get("rate", 0))
    targets = raw.get("targets")
    judge = raw.get("judge")
    if not 0 < rate <= 1:
        raise ValueError(f"queue {name}: shadow.rate must be in (0, 1]")
    if not isinstance(targets, list) or len(targets) < 1:
        raise ValueError(f"queue {name}: shadow.targets must list executors")
    for target in targets:
        kind, _, rest = str(target).partition(":")
        if kind not in _KINDS or not rest:
            raise ValueError(f"queue {name}: shadow target {target!r} is not kind:name")
        if kind == "openrouter" and not rest.endswith(":free"):
            raise ValueError(f"queue {name}: shadow openrouter targets must be :free endpoints")
    if not isinstance(judge, str) or not judge:
        raise ValueError(f"queue {name}: shadow.judge must name a task profile")
    return ShadowPolicy(
        rate, tuple(str(t) for t in targets), judge, int(raw.get("max_pending", 20))
    )
