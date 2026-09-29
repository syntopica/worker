"""POST /v1/jobs/{id}/ack."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.ack_result import ack_result
from worker.jobs.api_error import ApiError


def handle_ack(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Body ``{"result_id": ..., "decline": false}``."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    body = json.loads(ctx.body or b"{}")
    ack_result(
        ctx.conn,
        ctx.config,
        ctx.principal.name,
        ctx.parts[2],
        str(body.get("result_id")),
        bool(body.get("decline")),
        ctx.now,
    )
    return 200, {"acked": True}
