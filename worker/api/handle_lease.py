"""POST /v1/leases."""

import json
from typing import Any

from worker.api.request_context import RequestContext
from worker.jobs.api_error import ApiError
from worker.jobs.expire_leases import expire_leases
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest


def handle_lease(ctx: RequestContext) -> tuple[int, dict[str, Any]]:
    """200 with a lease, or 204 when nothing is eligible."""
    if ctx.principal.kind != "node":
        raise ApiError(403, "forbidden")
    body = json.loads(ctx.body or b"{}")
    if body.get("node") != ctx.principal.name:
        raise ApiError(403, "forbidden")
    expire_leases(ctx.conn, ctx.now)
    req = LeaseRequest(
        body["node"],
        body.get("resident_model"),
        bool(body["user_active"]),
        float(body["free_gb"]),
        float(body["current_idle_s"]),
    )
    lease = lease_job(ctx.conn, ctx.config, req, ctx.now)
    return (204, {}) if lease is None else (200, lease.to_json())
