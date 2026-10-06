"""Run one leased task under its profile and complete it, fenced by its attempt."""

import sys
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.attempt_report import attempt_report
from worker.node.lease_privacy_allowed import lease_privacy_allowed
from worker.node.report_shutdown import report_shutdown
from worker.tasks.run_task import run_task


def run_leased_task(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    lease: dict[str, Any],
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> None:
    """A running task is never preempted: its lease is heartbeated to the end."""
    attempt, generation = lease["attempt_id"], lease["generation"]
    profile = config.profiles.get(lease["model"])
    if profile is None or not lease_privacy_allowed(config, node_name, lease, "runner"):
        executor = {"node": node_name, "provider": "runner", "model": ""}
        refusal = "unknown_profile" if profile is None else "privacy_refused"
        report: dict[str, Any] | None = attempt_report("failed", executor, 0.0, error_code=refusal)
    else:
        started = clock()
        try:
            report = run_task(
                profile,
                lease,
                node_name,
                lambda: bool(link.heartbeat(attempt, generation, False)),
                clock,
                sleep,
            )
        except (SystemExit, KeyboardInterrupt):
            executor = {"node": node_name, "provider": profile.runner, "model": profile.model or ""}
            report_shutdown(link, attempt, generation, executor, clock() - started)
            raise
    if report is None:
        print(f"worker: task fenced job={lease.get('job_id')}", file=sys.stderr)
        return
    if report["outcome"] != "succeeded":
        # Outcome and error code are fixed allowlisted words; the job id is opaque.
        code = report.get("error_code")
        print(
            f"worker: task {report['outcome']}: {code} job={lease.get('job_id')}", file=sys.stderr
        )
    link.complete(attempt, generation, report)
