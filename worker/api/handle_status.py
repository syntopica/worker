"""GET /v1/status."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_status import read_status


def handle_status(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    return 200, read_status(ctx.conn, ctx.now)
