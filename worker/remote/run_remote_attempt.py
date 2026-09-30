"""Run one escalated inference job on OpenRouter, heartbeating while it runs."""

import threading
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.attempt_report import attempt_report
from worker.remote.openrouter_output import openrouter_output
from worker.remote.openrouter_request_body import openrouter_request_body
from worker.remote.post_openrouter import post_openrouter

_BEAT_S = 15.0
_CALL_TIMEOUT_S = 600.0


def run_remote_attempt(  # noqa: PLR0913, PLR0917
    lease: dict[str, Any],
    link: Any,
    node_name: str,
    key: str,
    config: WorkerConfig,
    clock: Callable[[], float],
    post: Callable[..., tuple[dict[str, Any] | None, str | None]] = post_openrouter,
    on_code: Callable[[str | None], None] | None = None,
) -> str:
    """Return the outcome; a fenced attempt is abandoned without completing.

    Non-public jobs go to zero-data-retention endpoints unless the queue's
    route turns that off. The call runs in a daemon thread so the lease is
    heartbeated; a fenced call is left to finish and its answer dropped.
    """
    queue = config.queues.get(str(lease.get("queue")))
    route = queue.openrouter if queue is not None else None
    zdr = (route.zdr if route is not None else True) and lease.get("privacy") != "public"
    body = openrouter_request_body(str(lease["model"]), lease["input"], zdr)
    executor = {"node": node_name, "provider": "openrouter", "model": str(lease["model"])}
    attempt, gen = lease["attempt_id"], lease["generation"]
    holder: dict[str, tuple[dict[str, Any] | None, str | None]] = {}
    thread = threading.Thread(
        target=lambda: holder.update(r=post(key, body, _CALL_TIMEOUT_S)), daemon=True
    )
    started = clock()
    thread.start()
    while True:
        thread.join(_BEAT_S)
        if not thread.is_alive():
            break
        if not link.heartbeat(attempt, gen, False):
            return "fenced"
    answer, error = holder.get("r", (None, "node_error"))
    if answer is None:
        if on_code is not None:
            on_code(error)
        report = attempt_report("failed", executor, clock() - started, error_code=error)
        link.complete(attempt, gen, report)
        return "failed"
    output, usage = openrouter_output(answer)
    report = attempt_report("succeeded", executor, clock() - started, output=output, usage=usage)
    link.complete(attempt, gen, report)
    return "succeeded"
