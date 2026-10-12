"""Unload this node's own models while the owner is using a local-idle-only node."""

from worker.node.node_memory import NodeMemory
from worker.node.unload_model import unload_model


def relieve_owner_active(
    url: str, memory: NodeMemory, resident: list[str], owner_active: bool
) -> list[str]:
    """Return the models unloaded because the owner is at the machine.

    A released attempt leaves the model resident for its whole ``keep_alive``,
    tens of GB that the owner's own work then competes with. On a node with
    ``local_idle_only`` the warm model buys nothing while the owner is active,
    so it goes as soon as they are. Only ``memory.owned`` models qualify, as
    for pressure and battery; a failed drain's model is left to
    ``retry_failed_drain``.

    One attempt per model per active spell: a failed unload is not retried
    every step, and an idle spell (which clears ``active_tried``) tries again.
    """
    if not owner_active:
        memory.active_tried.clear()
        return []
    mine = [
        m
        for m in resident
        if m in memory.owned and m != memory.failed_model and m not in memory.active_tried
    ]
    memory.active_tried.update(mine)
    done = [m for m in mine if unload_model(url, m)]
    memory.owned.difference_update(done)
    return done
