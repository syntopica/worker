"""Ask Ollama to drop a model from memory now."""

import json
import urllib.error
import urllib.request


def unload_model(url: str, model: str) -> bool:
    """``keep_alive: 0`` on /api/generate unloads after the current request."""
    request = urllib.request.Request(
        url + "/api/generate",
        json.dumps({"model": model, "keep_alive": 0}).encode(),
        {"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            response.read()
    except (urllib.error.URLError, TimeoutError, OSError):
        return False
    return True
