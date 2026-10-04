"""GET /v1/admin/audit?days=N."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_audit import read_audit

_DAY_S = 86400.0
_MAX_DAYS = 365.0


def handle_admin_audit(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; ``days`` defaults to 7 and is clamped to 1..365."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    days = max(1.0, min(float(ctx.query.get("days", 7)), _MAX_DAYS))
    return 200, {"rows": read_audit(ctx.conn, ctx.now - days * _DAY_S)}
