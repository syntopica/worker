"""The node loop: sample, report, lease, run - one attempt at a time."""

import time
from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.admission_block import admission_block
from worker.node.attempt_report import attempt_report
from worker.node.host_state import HostState
from worker.node.probe_quiet import probe_quiet
from worker.node.resident_models import resident_models
from worker.node.run_attempt import run_attempt
from worker.node.sample_host_state import sample_host_state
from worker.node.unload_model import unload_model

_REST_S = 30.0
_PRESSURE_BLOCKS = frozenset({"memory_pressure"})


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
    failed_model: str | None = None  # model of the lease that ended drain_failed
    unloaded = False  # unloaded once already in this pressure episode
    while True:
        state, now = sample(), clock()
        normal_since = (
            (normal_since if normal_since is not None else now)
            if state.pressure == "normal"
            else None
        )
        resident = resident_models(node.ollama_url) or []
        unexpected = [m for m in resident if m not in config.models]
        managed = next((m for m in resident if m in config.models), None)
        if failed_model is not None and probe_quiet(
            node.ollama_url, config.models[failed_model], timeout=20.0
        ):
            failed_model = None
        block = "drain_failed" if failed_model else admission_block(state, normal_since, now)
        if block is None:
            unloaded = False
        elif block in _PRESSURE_BLOCKS and managed is not None and not unloaded:
            unloaded = unload_model(node.ollama_url, config.models[managed].name)
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
        if lease is not None and lease["model"] not in config.models:
            executor = {"node": node.name, "provider": "ollama", "model": str(lease["model"])}
            report = attempt_report("failed", executor, 0.0, error_code="unknown_model")
            link.complete(lease["attempt_id"], lease["generation"], report)
        elif lease is not None:
            outcome = run_attempt(
                lease, link, config.models[lease["model"]], node, sample, sleep, clock
            )
            failed_model = lease["model"] if outcome == "drain_failed" else None
        elif forever:
            sleep(_REST_S)
        if not forever:
            return
