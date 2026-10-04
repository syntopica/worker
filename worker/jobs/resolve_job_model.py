"""The model, or the task profile, a submitted job runs on (spec 6, quality tiers)."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.check_task_grant import check_task_grant
from worker.jobs.submit_request import SubmitRequest
from worker.jobs.tier_route_for import tier_route_for


def resolve_job_model(config: WorkerConfig, req: SubmitRequest) -> str:
    """The tier route's choice comes first; raise ApiError when nothing usable is configured.

    A task's profile grant and privacy ceiling are checked here too.
    """
    route = tier_route_for(config.queues[req.queue], req.tier)
    if req.kind == "task":
        granted = check_task_grant(config, req)
        model: str | None = dict(route.profiles).get(granted, granted)
    else:
        preference = (*route.models, *req.models)
        model = next((m for m in preference if m in config.models), None)
    if model is None:
        raise ApiError(400, "unknown_model")
    return model
