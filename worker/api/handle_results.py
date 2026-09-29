"""GET /v1/results - long-poll without holding a transaction while waiting."""

import math
import time
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.expire_leases import expire_leases
from worker.jobs.list_results import list_results


def handle_results(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Return as soon as a result exists, or after ``wait`` seconds (max 30)."""
    if ctx.principal.kind != "producer":
        raise ApiError(403, "forbidden")
    queue = ctx.query.get("queue", "")
    if queue not in ctx.config.producers.get(ctx.principal.name, frozenset()):
        raise ApiError(403, "queue_not_granted")
    after, limit = int(ctx.query.get("after", 0)), int(ctx.query.get("limit", 50))
    wait = float(ctx.query.get("wait", 0))
    if not math.isfinite(wait):
        raise ApiError(400, "bad_request")
    deadline = time.monotonic() + max(0.0, min(wait, 30.0))
    while True:
        expire_leases(ctx.conn, time.time())
        rows = list_results(ctx.conn, ctx.principal.name, queue, after, limit)
        if rows or time.monotonic() >= deadline:
            return 200, {"results": rows}
        time.sleep(0.5)
