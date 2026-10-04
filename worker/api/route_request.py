"""Map (method, path shape) to its handler."""

from collections.abc import Callable
from typing import Any

from worker.api.handle_ack import handle_ack
from worker.api.handle_activity import handle_activity
from worker.api.handle_admin_ack import handle_admin_ack
from worker.api.handle_admin_audit import handle_admin_audit
from worker.api.handle_admin_cancel import handle_admin_cancel
from worker.api.handle_admin_content import handle_admin_content
from worker.api.handle_admin_job import handle_admin_job
from worker.api.handle_admin_jobs import handle_admin_jobs
from worker.api.handle_admin_retry import handle_admin_retry
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
from worker.api.path_matches import path_matches
from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError

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
    ("GET", "v1/activity"): handle_activity,
    ("GET", "v1/quality"): handle_quality,
    ("GET", "v1/admin/jobs"): handle_admin_jobs,
    ("GET", "v1/admin/jobs/*"): handle_admin_job,
    ("GET", "v1/admin/jobs/*/content"): handle_admin_content,
    ("POST", "v1/admin/jobs/*/cancel"): handle_admin_cancel,
    ("POST", "v1/admin/jobs/*/retry"): handle_admin_retry,
    ("POST", "v1/admin/jobs/*/ack"): handle_admin_ack,
    ("GET", "v1/admin/audit"): handle_admin_audit,
}


def route_request(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """A ``*`` in a pattern is one id segment; anything unmatched is 404.

    No two patterns of one method match the same path, so the first match is the only one.
    """
    for (method, pattern), handler in ROUTES.items():
        if method == ctx.method and path_matches(pattern.split("/"), ctx.parts):
            return handler(ctx)
    raise ApiError(404, "no_route")
