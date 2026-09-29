"""Wait until the backend is quiet: probe, then unload, then restart (spec 7)."""

from collections.abc import Callable

from worker.config.model_pin import ModelPin
from worker.node.probe_quiet import probe_quiet
from worker.node.restart_ollama import restart_ollama
from worker.node.unload_model import unload_model

_PROBE_S = 20.0


def drain_backend(  # noqa: PLR0913, PLR0917
    url: str,
    pin: ModelPin,
    label: str,
    heartbeat: Callable[[], bool],
    probe: Callable[..., bool] = probe_quiet,
    unload: Callable[[str, str], bool] = unload_model,
    restart: Callable[[str], bool] = restart_ollama,
) -> bool:
    """True once a probe answers; False is a drain failure and the node stays unavailable.

    A fenced heartbeat (False) stops the escalation after one probe.
    """
    for recover in (None, lambda: unload(url, pin.name), lambda: restart(label)):
        if recover is not None:
            recover()
        for _ in range(3):
            fenced = not heartbeat()
            if probe(url, pin, timeout=_PROBE_S):
                return True
            if fenced:
                return False  # fenced out: report the state, never unload or restart
    return False
