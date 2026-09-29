"""The models Ollama currently holds in memory."""

import http.client
import json
import urllib.error
import urllib.request


def resident_models(url: str) -> list[str] | None:
    """Names from /api/ps; None when the server does not answer."""
    try:
        with urllib.request.urlopen(url + "/api/ps", timeout=5) as response:
            return [m["name"] for m in json.load(response).get("models", [])]
    except (
        http.client.HTTPException,
        urllib.error.URLError,
        TimeoutError,
        OSError,
        ValueError,
        KeyError,
        TypeError,
        AttributeError,
    ):
        return None
