"""Check the source task's restrictions before changing its executor."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate


def on_demand_task_allowed(candidate: Candidate, config: WorkerConfig, node: str) -> bool:
    """Only ordinary tasks and judges may be overridden, within their source grant."""
    if candidate.kind != "task" or candidate.pin not in (None, "judge"):
        return False
    source = config.profiles.get(candidate.model)
    return (
        source is not None
        and not source.on_demand
        and candidate.privacy in source.privacy
        and (source.nodes is None or node in source.nodes)
    )
