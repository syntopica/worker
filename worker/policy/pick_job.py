"""Choose the next job: resident model, then priority, weighted share, age (spec 7)."""

from collections.abc import Mapping, Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest
from worker.policy.candidate_eligible import candidate_eligible


def pick_job(
    candidates: Sequence[Candidate],
    req: LeaseRequest,
    config: WorkerConfig,
    recent: Mapping[str, int],
    now: float,
) -> Candidate | None:
    """``recent`` counts attempts started per queue in the last hour."""
    eligible = [c for c in candidates if candidate_eligible(c, req, config)]
    if not eligible:
        return None
    resident = [c for c in eligible if c.model == req.resident_model]
    others = [c for c in eligible if c.model != req.resident_model]
    overdue = any(now - c.created > config.queues[c.queue].max_model_age_s for c in others)
    pool = resident if resident and not overdue else eligible
    top = max(c.priority for c in pool)
    tier = [c for c in pool if c.priority == top]
    return min(
        tier, key=lambda c: (recent.get(c.queue, 0) / config.queues[c.queue].weight, c.created)
    )
