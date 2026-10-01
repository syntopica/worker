"""One iteration of the OpenRouter loop: check headroom, lease, run."""

import sys
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.attempt_report import attempt_report
from worker.node.lease_privacy_allowed import lease_privacy_allowed
from worker.remote.openrouter_headroom import openrouter_headroom
from worker.remote.run_remote_attempt import run_remote_attempt

_NO_HEADROOM_REST_S = 900.0
_IDLE_REST_S = 30.0
_MIN_GAP_S = 3.1  # 20 requests per minute on free endpoints
_RATE_LIMITED_REST_S = 300.0


def remote_step(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    key: str,
    clock: Callable[[], float],
    headroom: Callable[[str], int | None] = openrouter_headroom,
    stopping: Callable[[], bool] = lambda: False,
) -> float:
    """Return the seconds to rest before the next iteration.

    Unknown headroom is no headroom (spec 8): the loop waits rather than
    spending a request it cannot account for.
    """
    left = headroom(key)
    if left is None or left <= 0:
        return _NO_HEADROOM_REST_S
    lease = link.lease_remote()
    if lease is None:
        return _IDLE_REST_S
    if not lease_privacy_allowed(config, node_name, lease, "openrouter"):
        executor = {"node": node_name, "provider": "openrouter", "model": str(lease["model"])}
        report = attempt_report("failed", executor, 0.0, error_code="privacy_refused")
        link.complete(lease["attempt_id"], lease["generation"], report)
        print(f"worker: remote failed: privacy_refused job={lease.get('job_id')}", file=sys.stderr)
        return _MIN_GAP_S
    codes: list[str | None] = []
    outcome = run_remote_attempt(
        lease, link, node_name, key, config, clock, on_code=codes.append, stopping=stopping
    )
    code = codes[-1] if codes else None
    if outcome != "succeeded":
        print(f"worker: remote {outcome}: {code} job={lease.get('job_id')}", file=sys.stderr)
    return _RATE_LIMITED_REST_S if code == "rate_limited" else _MIN_GAP_S
