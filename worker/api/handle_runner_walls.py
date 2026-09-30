"""POST /v1/nodes/{name}/walls - quota walls a node read from its own accounts."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError

_MAX_AHEAD_S = 40 * 86400.0


def handle_runner_walls(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Rest each named runner until its reported reset, never shortening a cooldown.

    The body maps a runner to an epoch reset; a reset in the past is ignored
    and one further than 40 days is refused as implausible.
    """
    if ctx.principal.kind != "node" or ctx.parts[2] != ctx.principal.name:
        raise ApiError(403, "forbidden")
    walls = json.loads(ctx.body)
    if not isinstance(walls, dict):
        raise ApiError(400, "bad_request")
    for runner, until in walls.items():
        if not isinstance(runner, str) or not isinstance(until, int | float):
            raise ApiError(400, "bad_request")
        if until > ctx.now + _MAX_AHEAD_S:
            raise ApiError(400, "bad_request")
        if until > ctx.now:
            ctx.conn.execute(
                "INSERT INTO cooldowns (runner, until) VALUES (?, ?)"
                " ON CONFLICT(runner) DO UPDATE SET until=max(until, excluded.until)",
                (runner, float(until)),
            )
    return 200, {"ok": True}
