"""Unload this node's own models while the host runs on battery."""

from worker.node.node_memory import NodeMemory
from worker.node.unload_model import unload_model


def relieve_battery(
    url: str, memory: NodeMemory, resident: list[str], on_ac: bool | None
) -> list[str]:
    """Return the models unloaded because the host is on battery.

    A battery release drains the attempt but leaves the model resident for its
    whole ``keep_alive``, holding tens of GB on a discharging laptop. Unlike a
    short ``user_active`` gap, battery lasts, so the warm model buys nothing.
    Only ``memory.owned`` models qualify, as for pressure; an unknown power
    state unloads nothing. A failed drain's model is left to
    ``retry_failed_drain``, whose timer would otherwise be bypassed every step.

    One attempt per model per battery episode: a failed unload is not retried
    every step against an unreachable backend, and it keeps its ownership, so
    the next episode (AC clears ``battery_tried``) tries again.
    """
    if on_ac is True:
        memory.battery_tried.clear()
    if on_ac is not False:
        return []
    mine = [
        m
        for m in resident
        if m in memory.owned and m != memory.failed_model and m not in memory.battery_tried
    ]
    memory.battery_tried.update(mine)
    done = [m for m in mine if unload_model(url, m)]
    memory.owned.difference_update(done)
    return done
