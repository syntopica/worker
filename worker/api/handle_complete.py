"""POST /v1/attempts/{id}/complete."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport


def handle_complete(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Return the job's new state."""
    if ctx.principal.kind != "node":
        raise ApiError(403, "forbidden")
    b = json.loads(ctx.body or b"{}")
    report = CompletionReport(
        b["outcome"],
        b.get("output"),
        b.get("usage") or {},
        b.get("executor") or {},
        b.get("error_code"),
        float(b.get("wall_s", 0)),
    )
    return 200, {
        "state": complete_attempt(
            ctx.conn, ctx.config, ctx.parts[2], int(b["generation"]), report, ctx.now
        )
    }
