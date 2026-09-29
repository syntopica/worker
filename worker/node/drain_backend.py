"""Wait until the backend is quiet: probe, then unload, then restart (spec 7)."""

import time
from collections.abc import Callable

from worker.config.model_pin import ModelPin
from worker.node.probe_quiet import probe_quiet
from worker.node.resident_models import resident_models
from worker.node.restart_ollama import restart_ollama
from worker.node.unload_model import unload_model
from worker.node.wait_unlisted import wait_unlisted

_PROBE_S = 20.0
_PAUSE_S = 2.0


def drain_backend(  # noqa: PLR0913, PLR0917
    url: str,
    pin: ModelPin,
    label: str,
    heartbeat: Callable[[], bool],
    probe: Callable[..., bool] = probe_quiet,
    unload: Callable[[str, str], bool] = unload_model,
    restart: Callable[[str], bool] = restart_ollama,
    resident: Callable[[str], list[str] | None] = resident_models,
    pause: Callable[[float], None] = time.sleep,
) -> bool:
    """True once the backend is confirmed quiet; False is a drain failure.

    1. Up to three one-token probes, each only while the model is resident
       (a probe of an absent model would load it). A model /api/ps no longer
       lists is quiet. A fenced heartbeat stops here: never unload or restart.
    2. Unload; quiet once /api/ps stops listing the model. No probe follows.
    3. ``restart_ollama`` (kickstart) only when the model is still resident
       after the unload, and therefore still busy: every probe timed out.
       An unreachable /api/ps never triggers a restart.
    A heartbeat runs between every step, so a long drain keeps its lease.
    """
    for _ in range(3):
        fenced = not heartbeat()
        listed = resident(url)
        if listed is not None and pin.name not in listed:
            return True
        if listed is not None and probe(url, pin, timeout=_PROBE_S):
            return True
        if fenced:
            return False
        if listed is None:
            pause(_PAUSE_S)
    heartbeat()
    unload(url, pin.name)
    listed = wait_unlisted(url, pin.name, heartbeat, resident, pause)
    if listed is None or pin.name not in listed:
        return listed is not None
    heartbeat()
    restart(label)
    listed = wait_unlisted(url, pin.name, heartbeat, resident, pause)
    return listed is not None and pin.name not in listed
