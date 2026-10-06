"""The queued inference an on-demand profile may take for the asking node."""

import dataclasses
from collections.abc import Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows
from worker.policy.profile_resting import profile_resting


def on_demand_candidates(
    candidates: Sequence[Candidate],
    config: WorkerConfig,
    cooling: frozenset[str],
    node: tuple[str, str],
    name: str,
) -> list[Candidate]:
    """Unpinned inference jobs, carrying the profile in ``model``, from any queue.

    The profile must be on demand, rested and allowed on the node; each job's
    class must be allowed by the profile and for ``runner`` executors. Tasks
    and pinned jobs stay with the executors they were submitted or pinned to.
    ``node`` is the asking node's name and trust class.
    """
    profile = config.profiles.get(name)
    if profile is None or not profile.on_demand or profile_resting(profile, cooling):
        return []
    if profile.nodes is not None and node[0] not in profile.nodes:
        return []
    return [
        dataclasses.replace(c, model=name)
        for c in candidates
        if c.kind == "inference"
        and c.pin is None
        and c.privacy in profile.privacy
        and privacy_allows(config, c.privacy, "runner", node[1])
    ]
