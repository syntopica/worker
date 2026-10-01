"""Build one QueuePolicy from its raw configuration entry."""

from typing import Any

from worker.config.parse_openrouter_route import parse_openrouter_route
from worker.config.parse_runner_route import parse_runner_route
from worker.config.parse_shadow_policy import parse_shadow_policy
from worker.config.parse_tier_routes import parse_tier_routes
from worker.config.queue_policy import QueuePolicy

_RUN_WHEN = ("idle", "active_ok")


def parse_queue_policy(name: str, raw: dict[str, Any]) -> QueuePolicy:
    """Apply the spec defaults; raise ValueError on an unknown ``run_when``.

    ``retention_days`` must cover ``unacked_ttl_hours``: a shorter one would
    sweep a failed control result before its producer fetched it.
    """
    run_when = raw.get("run_when", "idle")
    if run_when not in _RUN_WHEN:
        raise ValueError(f"queue {name}: run_when must be one of {_RUN_WHEN}")
    retention_days = int(raw.get("retention_days", 7))
    unacked_ttl_hours = int(raw.get("unacked_ttl_hours", 72))
    if retention_days < 1 or retention_days * 24 < unacked_ttl_hours:
        raise ValueError(f"queue {name}: retention_days must be >= 1 and cover unacked_ttl_hours")
    return QueuePolicy(
        name=name,
        run_when=run_when,
        weight=int(raw.get("weight", 1)),
        max_outstanding=int(raw.get("max_outstanding", 200)),
        retention_days=retention_days,
        unacked_ttl_hours=unacked_ttl_hours,
        parked_min_idle_s=float(raw.get("parked_min_idle_s", 600)),
        max_model_age_s=float(raw.get("max_model_age_s", 3600)),
        profiles=frozenset(raw.get("profiles", ())),
        fallbacks=tuple(
            (str(primary), tuple(str(a) for a in alternatives))
            for primary, alternatives in (raw.get("fallbacks") or {}).items()
        ),
        openrouter=parse_openrouter_route(name, raw.get("openrouter")),
        tiers=parse_tier_routes(name, raw.get("tiers")),
        runner=parse_runner_route(name, raw.get("runner")),
        local_after_s=float(raw.get("local_after_s", 0)),
        shadow=parse_shadow_policy(name, raw.get("shadow")),
    )
