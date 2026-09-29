"""The node loop: sample, report, lease, run - one attempt at a time."""

import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.admission_block import admission_block
from worker.node.host_state import HostState
from worker.node.probe_quiet import probe_quiet
from worker.node.resident_models import resident_models
from worker.node.run_attempt import run_attempt
from worker.node.sample_host_state import sample_host_state

_REST_S = 30.0


def run_node(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    sample: Callable[[], HostState] = sample_host_state,
    sleep: Callable[[float], None] = time.sleep,
    clock: Callable[[], float] = time.time,
    forever: bool = True,
) -> None:
    """Never takes work while a previous drain failed and the backend is still busy."""
    node = config.nodes[node_name]
    normal_since: float | None = None
    drain_failed = False
    while True:
        state, now = sample(), clock()
        normal_since = (normal_since or now) if state.pressure == "normal" else None
        resident = resident_models(node.ollama_url) or []
        unexpected = [m for m in resident if m not in config.models]
        managed = next((m for m in resident if m in config.models), None)
        if drain_failed and managed is not None:
            drain_failed = not probe_quiet(node.ollama_url, config.models[managed], timeout=20.0)
        block = "drain_failed" if drain_failed else admission_block(state, normal_since, now)
        link.report(
            {
                "reason": block,
                "idle_s": state.idle_s,
                "on_ac": state.on_ac,
                "pressure": state.pressure,
                "resident": resident,
                "unexpected": unexpected,
            }
        )
        lease = None
        if block is None:
            footprint = config.models[managed].cold_gb if managed else 0.0
            user_active = (state.idle_s or 0.0) < node.idle_threshold_s
            free_gb = node.memory_budget_gb - footprint
            lease = link.lease(managed, user_active, free_gb, state.idle_s or 0.0)
        if lease is not None:
            outcome = run_attempt(
                lease, link, config.models[lease["model"]], node, sample, sleep, clock
            )
            drain_failed = outcome == "drain_failed"
        elif forever:
            sleep(_REST_S)
        if not forever:
            return
