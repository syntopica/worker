"""GET /v1/jobs/{id}."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.get_job import get_job
from worker.jobs.job_cooling_until import job_cooling_until


def handle_get_job(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """The job's state, its latest unacknowledged result and any cooldown holding it."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    job = get_job(ctx.conn, ctx.principal.name, ctx.parts[2])
    job["cooling_until"] = job_cooling_until(ctx.conn, ctx.config, job["id"], ctx.now)
    return 200, job
