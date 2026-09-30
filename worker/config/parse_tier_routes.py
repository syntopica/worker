"""Build a queue's tier routes from its raw ``tiers`` entry."""

from typing import Any

from worker.config.tier_route import TierRoute
from worker.jobs.states import TIERS


def parse_tier_routes(queue: str, raw: Any) -> tuple[tuple[str, TierRoute], ...]:
    """Raise ValueError on an unknown tier or a malformed entry; ``None`` means no tiers."""
    if raw is None:
        return ()
    if not isinstance(raw, dict):
        raise ValueError(f"queue {queue}: tiers must be an object")
    routes: list[tuple[str, TierRoute]] = []
    for tier, entry in raw.items():
        if tier not in TIERS or not isinstance(entry, dict):
            raise ValueError(f"queue {queue}: tiers.{tier} must be one of {TIERS} and an object")
        models = entry.get("models", [])
        profiles = entry.get("profiles", {})
        if not isinstance(models, list) or not isinstance(profiles, dict):
            raise ValueError(f"queue {queue}: tiers.{tier} needs a models list and a profiles map")
        routes.append(
            (
                tier,
                TierRoute(
                    models=tuple(str(m) for m in models),
                    profiles=tuple((str(k), str(v)) for k, v in profiles.items()),
                ),
            )
        )
    return tuple(routes)
