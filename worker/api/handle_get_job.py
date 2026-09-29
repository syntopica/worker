"""GET /v1/jobs/{id}."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.get_job import get_job


def handle_get_job(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """The job's state and its latest unacknowledged result."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    return 200, get_job(ctx.conn, ctx.principal.name, ctx.parts[2])
