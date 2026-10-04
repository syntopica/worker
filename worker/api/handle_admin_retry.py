"""POST /v1/admin/jobs/{id}/retry."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.retry_job import retry_job


def handle_admin_retry(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """201 with the new job, 200 when this original was already retried."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    original = ctx.parts[3]
    new_id, created = retry_job(ctx.conn, ctx.config, original, ctx.principal.name, ctx.now)
    state = ctx.conn.execute("SELECT state FROM jobs WHERE id=?", (new_id,)).fetchone()[0]
    return (201 if created else 200), {"id": new_id, "state": state, "retry_of": original}
