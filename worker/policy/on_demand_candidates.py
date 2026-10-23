"""The queued inference an on-demand profile may take for the asking node."""

import dataclasses
from collections.abc import Sequence

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.on_demand_inference_allowed import on_demand_inference_allowed
from worker.policy.on_demand_task_allowed import on_demand_task_allowed
from worker.policy.privacy_allows import privacy_allows
from worker.policy.profile_resting import profile_resting


def on_demand_candidates(  # noqa: PLR0913, PLR0917
    candidates: Sequence[Candidate],
    config: WorkerConfig,
    cooling: frozenset[str],
    node: tuple[str, str],
    name: str,
    override_pins: bool = False,
) -> list[Candidate]:
    """Inference, prompt-only tasks and judges under an explicit execution profile.

    The profile must be on demand, rested and allowed on the node; each job's
    class must be allowed by the profile and for ``runner`` executors.
    Comparison pins stay with their submitted executors.
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
        if (
            on_demand_inference_allowed(c, override_pins)
            or on_demand_task_allowed(c, config, node[0])
        )
        and c.privacy in profile.privacy
        and privacy_allows(config, c.privacy, "runner", node[1])
    ]
