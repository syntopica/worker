"""Map (method, path shape) to its handler."""

from collections.abc import Callable
from typing import Any

from worker.api.handle_ack import handle_ack
from worker.api.handle_cancel import handle_cancel
from worker.api.handle_complete import handle_complete
from worker.api.handle_costs import handle_costs
from worker.api.handle_get_job import handle_get_job
from worker.api.handle_heartbeat import handle_heartbeat
from worker.api.handle_lease import handle_lease
from worker.api.handle_node_report import handle_node_report
from worker.api.handle_quality import handle_quality
from worker.api.handle_results import handle_results
from worker.api.handle_runner_walls import handle_runner_walls
from worker.api.handle_status import handle_status
from worker.api.handle_submit import handle_submit
from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError

_WILDCARD_INDEX = 2
Handler = Callable[[RequestContext], tuple[int, dict[str, Any]]]
ROUTES: dict[tuple[str, str], Handler] = {
    ("POST", "v1/jobs"): handle_submit,
    ("GET", "v1/jobs/*"): handle_get_job,
    ("POST", "v1/jobs/*/cancel"): handle_cancel,
    ("POST", "v1/jobs/*/ack"): handle_ack,
    ("GET", "v1/results"): handle_results,
    ("POST", "v1/leases"): handle_lease,
    ("POST", "v1/attempts/*/heartbeat"): handle_heartbeat,
    ("POST", "v1/attempts/*/complete"): handle_complete,
    ("POST", "v1/nodes/*/report"): handle_node_report,
    ("POST", "v1/nodes/*/walls"): handle_runner_walls,
    ("GET", "v1/status"): handle_status,
    ("GET", "v1/costs"): handle_costs,
    ("GET", "v1/quality"): handle_quality,
}


def route_request(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """The third path segment is the wildcard; anything unmatched is 404."""
    shape = "/".join("*" if i == _WILDCARD_INDEX else p for i, p in enumerate(ctx.parts))
    handler = ROUTES.get((ctx.method, shape))
    if handler is None:
        raise ApiError(404, "no_route")
    return handler(ctx)
