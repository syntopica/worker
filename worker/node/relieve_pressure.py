"""Unload this node's own models once memory pressure is sustained."""

from worker.node.node_memory import NodeMemory
from worker.node.pressure_samples import PRESSURE_SAMPLES
from worker.node.unload_model import unload_model


def relieve_pressure(url: str, memory: NodeMemory, resident: list[str]) -> list[str]:
    """Once per pressure episode; return the models unloaded.

    Only ``memory.owned`` models qualify: never one resident at startup or
    loaded by another application.
    """
    if memory.pressure_streak < PRESSURE_SAMPLES or memory.unloaded:
        return []
    mine = [m for m in resident if m in memory.owned]
    done = [m for m in mine if unload_model(url, m)]
    memory.owned.difference_update(done)
    memory.unloaded = bool(mine)
    return done
