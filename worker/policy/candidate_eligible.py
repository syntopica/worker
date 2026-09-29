"""Whether one candidate may run on the requesting node right now (spec 7)."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest


def candidate_eligible(c: Candidate, req: LeaseRequest, config: WorkerConfig) -> bool:
    """Queue run policy, incremental memory and parking; privacy is filtered earlier."""
    queue = config.queues.get(c.queue)
    if queue is None:
        return False  # a queue removed from the configuration runs nothing
    if req.user_active and queue.run_when != "active_ok":
        return False
    pin = config.models.get(c.model)
    if pin is None:
        return False
    budget = config.nodes[req.node].memory_budget_gb
    fits = pin.warm_gb <= req.free_gb if c.model == req.resident_model else pin.cold_gb <= budget
    if not fits:
        return False
    return not (c.parked and req.current_idle_s < c.parked_min_idle_s)
