"""POST /v1/admin/jobs/{id}/ack."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.admin_ack_job import admin_ack_job
from worker.jobs.api_error import ApiError
from worker.jobs.load_admin_job import load_admin_job
from worker.jobs.record_audit import record_audit


def handle_admin_ack(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Acknowledge a failed or control result for its producer. Audited."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    job = load_admin_job(ctx.conn, ctx.parts[3])
    admin_ack_job(ctx.conn, ctx.config, job, ctx.now)
    record_audit(ctx.conn, "ack", job["id"], job["privacy"], ctx.principal.name, ctx.now)
    state = ctx.conn.execute("SELECT state FROM jobs WHERE id=?", (job["id"],)).fetchone()[0]
    return 200, {"id": job["id"], "state": state}
