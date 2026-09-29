"""Confirm the backend has stopped the previous request (spec 7, amendment 2)."""

import json
import urllib.error
import urllib.request

from worker.config.model_pin import ModelPin


def probe_quiet(url: str, pin: ModelPin, timeout: float) -> bool:
    """A one-token request with the pinned options; with NUM_PARALLEL=1 it waits its turn."""
    body = {
        "model": pin.name,
        "messages": [{"role": "user", "content": "."}],
        "stream": False,
        "think": False,
        "keep_alive": pin.keep_alive,
        "options": {"num_ctx": pin.num_ctx, "num_predict": 1},
    }
    request = urllib.request.Request(
        url + "/api/chat", json.dumps(body).encode(), {"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            response.read()
    except (urllib.error.URLError, TimeoutError, OSError):
        return False
    return True
