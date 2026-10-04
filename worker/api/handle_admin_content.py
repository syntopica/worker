"""GET /v1/admin/jobs/{id}/content."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.load_admin_job import load_admin_job
from worker.jobs.load_job_input import load_job_input
from worker.jobs.load_job_output import load_job_output
from worker.jobs.record_audit import record_audit
from worker.jobs.states import SENSITIVE


def handle_admin_content(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """A sensitive class needs ``X-Worker-Reveal`` naming it, and every such read is audited."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    job = load_admin_job(ctx.conn, ctx.parts[3])
    sensitive = job["privacy"] in SENSITIVE
    if sensitive and ctx.reveal != job["privacy"]:
        raise ApiError(403, "reveal_required")
    content = {
        "input": load_job_input(ctx.conn, job["id"]),
        "output": load_job_output(ctx.conn, job["id"]),
    }
    if content["input"] is None and content["output"] is None:
        raise ApiError(410, "content_gone")
    if sensitive:
        record_audit(ctx.conn, "reveal", job["id"], job["privacy"], ctx.principal.name, ctx.now)
    return 200, content
