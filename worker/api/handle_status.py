"""GET /v1/status."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.openrouter_headroom import openrouter_headroom
from worker.jobs.read_status import read_status


def handle_status(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; ``openrouter`` is today's use of the free key."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    status = read_status(ctx.conn, ctx.now)
    status["openrouter"] = openrouter_headroom(ctx.config, ctx.conn, ctx.now)
    return 200, status
