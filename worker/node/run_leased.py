"""Run one granted lease; an unexpected error becomes a node_error report."""

import sys
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.attempt_report import attempt_report
from worker.node.host_state import HostState
from worker.node.node_memory import NodeMemory
from worker.node.run_attempt import run_attempt
from worker.node.settle_attempt import settle_attempt


def run_leased(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    memory: NodeMemory,
    lease: dict[str, Any],
    resident: list[str],
    sample: Callable[[], HostState],
    sleep: Callable[[float], None],
    clock: Callable[[], float],
) -> bool:
    """Return True when the node should rest before its next lease."""
    node = config.nodes[node_name]
    model = str(lease["model"])
    executor = {"node": node.name, "provider": "ollama", "model": model}
    if model not in config.models:
        report = attempt_report("failed", executor, 0.0, error_code="unknown_model")
        link.complete(lease["attempt_id"], lease["generation"], report)
        return False
    codes: list[str | None] = [None]
    try:
        outcome = run_attempt(
            lease, link, config.models[model], node, sample, sleep, clock, codes.append
        )
    except Exception as error:
        print(f"worker: attempt failed: {type(error).__name__}", file=sys.stderr)
        report = attempt_report("failed", executor, 0.0, error_code="node_error")
        link.complete(lease["attempt_id"], lease["generation"], report)
        outcome, codes[-1:] = "failed", ["node_error"]
    return settle_attempt(memory, model, model in resident, outcome, codes[-1], clock())
