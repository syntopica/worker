"""GET /v1/admin/jobs/{id}."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_admin_job import read_admin_job


def handle_admin_job(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; metadata, attempts and payload presence, never content."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    return 200, read_admin_job(ctx.conn, ctx.parts[3])
