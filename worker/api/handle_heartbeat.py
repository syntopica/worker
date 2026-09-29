"""POST /v1/attempts/{id}/heartbeat."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.heartbeat_attempt import heartbeat_attempt


def handle_heartbeat(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """409 tells the node its attempt no longer owns the job."""
    if ctx.principal.kind != "node":
        raise ApiError(403, "forbidden")
    body = json.loads(ctx.body or b"{}")
    heartbeat_attempt(
        ctx.conn, ctx.parts[2], int(body["generation"]), bool(body.get("draining")), ctx.now
    )
    return 200, {"ok": True}
