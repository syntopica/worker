"""GET /v1/costs?days=N."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_costs import read_costs

_DAY_S = 86400.0


def handle_costs(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; ``days`` defaults to 7."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    days = float(ctx.query.get("days", 7))
    return 200, {"rows": read_costs(ctx.conn, ctx.now - days * _DAY_S)}
