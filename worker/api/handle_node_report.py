"""POST /v1/nodes/{name}/report - the node's host state for `worker status`."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError


def handle_node_report(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Store the report verbatim; it carries metadata only (reason, model, idle)."""
    if ctx.principal.kind != "node" or ctx.parts[2] != ctx.principal.name:
        raise ApiError(403, "forbidden")
    ctx.conn.execute(
        "INSERT INTO nodes (name, report, updated) VALUES (?, ?, ?)"
        " ON CONFLICT(name) DO UPDATE SET report=excluded.report, updated=excluded.updated",
        (ctx.principal.name, ctx.body.decode(), ctx.now),
    )
    return 200, {"ok": True}
