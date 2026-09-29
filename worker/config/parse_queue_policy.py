"""Build one QueuePolicy from its raw configuration entry."""

from typing import Any

from worker.config.queue_policy import QueuePolicy

_RUN_WHEN = ("idle", "active_ok")


def parse_queue_policy(name: str, raw: dict[str, Any]) -> QueuePolicy:
    """Apply the spec defaults; raise ValueError on an unknown ``run_when``."""
    run_when = raw.get("run_when", "idle")
    if run_when not in _RUN_WHEN:
        raise ValueError(f"queue {name}: run_when must be one of {_RUN_WHEN}")
    return QueuePolicy(
        name=name,
        run_when=run_when,
        weight=int(raw.get("weight", 1)),
        max_outstanding=int(raw.get("max_outstanding", 200)),
        retention_days=int(raw.get("retention_days", 7)),
        unacked_ttl_hours=int(raw.get("unacked_ttl_hours", 72)),
        parked_min_idle_s=float(raw.get("parked_min_idle_s", 600)),
        max_model_age_s=float(raw.get("max_model_age_s", 3600)),
    )
