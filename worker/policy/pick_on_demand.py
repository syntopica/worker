"""Choose the next job for an on-demand profile: priority, weighted share, age."""

from collections.abc import Mapping, Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate


def pick_on_demand(
    candidates: Sequence[Candidate], config: WorkerConfig, recent: Mapping[str, int]
) -> Candidate | None:
    """``pick_task``'s ordering over candidates already filtered for the profile.

    No run policy: the owner asked for this run, so an active user holds nothing back.
    """
    if not candidates:
        return None
    top = max(c.priority for c in candidates)
    tier = [c for c in candidates if c.priority == top]
    return min(
        tier, key=lambda c: (recent.get(c.queue, 0) / config.queues[c.queue].weight, c.created)
    )
