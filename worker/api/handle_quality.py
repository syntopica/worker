"""GET /v1/quality?days=N."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.read_attempt_quality import read_attempt_quality
from worker.jobs.read_judgements import read_judgements
from worker.jobs.read_result_ratings import read_result_ratings

_DAY_S = 86400.0


def handle_quality(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; ``days`` defaults to 7."""
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    since = ctx.now - float(ctx.query.get("days", 7)) * _DAY_S
    return 200, {
        "attempts": read_attempt_quality(ctx.conn, since),
        "ratings": read_result_ratings(ctx.conn, since),
        "judged": read_judgements(ctx.conn, since),
    }
