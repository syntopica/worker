"""One iteration of the node loop: sample, report, and maybe lease and run."""

from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.node.host_state import HostState
from worker.node.node_block import node_block
from worker.node.node_memory import NodeMemory
from worker.node.observe_host_state import observe_host_state
from worker.node.relieve_pressure import relieve_pressure
from worker.node.resident_models import resident_models
from worker.node.run_leased import run_leased


def node_step(  # noqa: PLR0913, PLR0917
    config: WorkerConfig,
    node_name: str,
    link: Any,
    memory: NodeMemory,
    sample: Callable[[], HostState],
    sleep: Callable[[float], None],
    clock: Callable[[], float],
) -> bool:
    """Return True when the loop should rest before the next iteration.

    A failed drain clears once /api/ps no longer lists its model, checked
    only while pressure is normal; no chat probe is ever sent for it.
    """
    node = config.nodes[node_name]
    state, now = sample(), clock()
    observe_host_state(memory, state, now)
    answer = resident_models(node.ollama_url)
    resident = answer or []
    if answer is not None:
        memory.owned.intersection_update(resident)
        if state.pressure == "normal" and memory.failed_model not in resident:
            memory.failed_model = None
    relieve_pressure(node.ollama_url, memory, resident)
    block = node_block(memory, state, now, answer is None)
    link.report(
        {
            "reason": block,
            "idle_s": state.idle_s,
            "on_ac": state.on_ac,
            "pressure": state.pressure,
            "resident": resident,
            "unexpected": [m for m in resident if m not in config.models],
        }
    )
    if block is not None:
        return True
    managed = next((m for m in resident if m in config.models), None)
    footprint = config.models[managed].cold_gb if managed else 0.0
    idle = state.idle_s or 0.0
    lease = link.lease(
        managed, idle < node.idle_threshold_s, node.memory_budget_gb - footprint, idle
    )
    if lease is None:
        return True
    return run_leased(config, node_name, link, memory, lease, resident, sample, sleep, clock)
