"""POST /v1/admin/jobs/{id}/cancel."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.cancel_job import cancel_job
from worker.jobs.load_admin_job import load_admin_job
from worker.jobs.record_audit import record_audit


def handle_admin_cancel(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Any producer's job; a terminal job keeps its state. Audited."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    job = load_admin_job(ctx.conn, ctx.parts[3])
    state = cancel_job(ctx.conn, None, job["id"], ctx.now)
    record_audit(ctx.conn, "cancel", job["id"], job["privacy"], ctx.principal.name, ctx.now)
    return 200, {"id": job["id"], "state": state}
