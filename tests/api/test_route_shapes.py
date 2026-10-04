import pytest

from worker.api import route_request as routing
from worker.api.path_matches import path_matches
from worker.api.request_context import RequestContext
from worker.auth.principal import Principal
from worker.jobs.api_error import ApiError


@pytest.mark.parametrize(
    ("pattern", "path", "matched"),
    [
        ("v1/jobs/*", "v1/jobs/abc", True),
        ("v1/jobs/*/cancel", "v1/jobs/abc/cancel", True),
        ("v1/admin/jobs/*", "v1/admin/jobs/abc", True),
        ("v1/admin/jobs/*/content", "v1/admin/jobs/abc/content", True),
        ("v1/admin/jobs", "v1/admin/jobs", True),
        ("v1/jobs/*", "v1/admin/jobs", False),
        ("v1/admin/jobs/*", "v1/admin/jobs/abc/content", False),
        ("v1/admin/jobs/*/cancel", "v1/admin/jobs/abc/retry", False),
    ],
)
def test_a_star_matches_exactly_one_segment(pattern, path, matched):
    assert path_matches(pattern.split("/"), path.split("/")) is matched


def ctx(method, path):
    return RequestContext(
        Principal("admin", "admin"), method, path.split("/"), {}, b"", None, None, 0.0
    )


def test_admin_paths_reach_their_handlers(monkeypatch):
    seen = []
    for key in list(routing.ROUTES):
        monkeypatch.setitem(routing.ROUTES, key, lambda c, key=key: (200, seen.append(key) or {}))
    for method, path in (
        ("GET", "v1/admin/jobs"),
        ("GET", "v1/admin/jobs/abc"),
        ("GET", "v1/admin/jobs/abc/content"),
        ("POST", "v1/admin/jobs/abc/cancel"),
        ("POST", "v1/admin/jobs/abc/retry"),
        ("POST", "v1/admin/jobs/abc/ack"),
        ("GET", "v1/admin/audit"),
        ("GET", "v1/jobs/abc"),
    ):
        routing.route_request(ctx(method, path))
    assert len(set(seen)) == 8
    with pytest.raises(ApiError) as refused:
        routing.route_request(ctx("POST", "v1/admin/jobs/abc"))
    assert refused.value.code == "no_route"
