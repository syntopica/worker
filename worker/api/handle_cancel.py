"""POST /v1/jobs/{id}/cancel."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.cancel_job import cancel_job


def handle_cancel(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Return the state the job ended in."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    return 200, {"state": cancel_job(ctx.conn, ctx.principal.name, ctx.parts[2], ctx.now)}
