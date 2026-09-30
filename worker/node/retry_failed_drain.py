"""Retry the unload of a model whose drain failed, on a slow timer."""

from collections.abc import Callable

from worker.node.node_memory import NodeMemory
from worker.node.unload_model import unload_model

_RETRY_S = 300.0


def retry_failed_drain(
    url: str,
    memory: NodeMemory,
    resident: list[str],
    now: float,
    unload: Callable[[str, str], bool] = unload_model,
) -> bool:
    """Return True when an unload was sent.

    A drain whose unload and restart both failed leaves the model listed and
    the node blocked; nothing else would try again. Once every five minutes,
    while the model is still listed, send another unload. The block itself
    clears in ``node_step`` when /api/ps stops listing the model.
    """
    model = memory.failed_model
    if model is None or model not in resident or now < memory.drain_retry_at:
        return False
    memory.drain_retry_at = now + _RETRY_S
    unload(url, model)
    return True
