"""POST /v1/jobs."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.submit_job import submit_job


def handle_submit(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """201 for a new job, 200 when the idempotency key matched an existing one."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    job_id, created = submit_job(ctx.conn, ctx.config, ctx.principal.name, ctx.body, ctx.now)
    return (201 if created else 200), {"id": job_id, "created": created}
