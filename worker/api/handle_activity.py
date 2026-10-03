"""GET /v1/activity?hours=N."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_activity import read_activity


def handle_activity(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; ``hours`` defaults to 24."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    try:
        hours = float(ctx.query.get("hours", 24))
    except ValueError as error:
        raise ApiError(400, "bad_request") from error
    return 200, read_activity(ctx.conn, ctx.now, hours)
