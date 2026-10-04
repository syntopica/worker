"""GET /v1/admin/jobs?queue=&state=&producer=&before=&limit=."""

from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.jobs_filter import JobsFilter
from worker.jobs.list_admin_jobs import list_admin_jobs
from worker.jobs.states import JOB_STATES

_MAX_LIMIT = 100


def handle_admin_jobs(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """Admin only; a queue must be configured and every state known, ``limit`` is clamped.

    ``state`` is one state or a comma-separated list of them:
    ``leased,running,draining`` is what is running now.
    """
    if ctx.principal.kind != "admin":
        raise ApiError(403, "forbidden")
    queue, state = ctx.query.get("queue"), ctx.query.get("state")
    if queue is not None and queue not in ctx.config.queues:
        raise ApiError(400, "unknown_queue")
    states = tuple(state.split(",")) if state is not None else None
    if states is not None and not all(s in JOB_STATES for s in states):
        raise ApiError(400, "unknown_state")
    wanted = JobsFilter(
        queue=queue,
        states=states,
        producer=ctx.query.get("producer"),
        before=ctx.query.get("before"),
        limit=max(1, min(int(ctx.query.get("limit", 50)), _MAX_LIMIT)),
    )
    return 200, list_admin_jobs(ctx.conn, wanted)
