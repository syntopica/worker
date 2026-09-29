"""POST /v1/nodes/{name}/report - the node's host state for `worker status`."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError


def handle_node_report(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Store the report re-serialised (it must be a JSON object); it carries metadata only (reason, model, idle)."""
    if ctx.principal.kind != "node" or ctx.parts[2] != ctx.principal.name:
        raise ApiError(403, "forbidden")
    report = json.loads(ctx.body)
    if not isinstance(report, dict):
        raise ApiError(400, "bad_request")
    ctx.conn.execute(
        "INSERT INTO nodes (name, report, updated) VALUES (?, ?, ?)"
        " ON CONFLICT(name) DO UPDATE SET report=excluded.report, updated=excluded.updated",
        (ctx.principal.name, json.dumps(report), ctx.now),
    )
    return 200, {"ok": True}
