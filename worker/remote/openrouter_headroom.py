"""Free-model requests left today, as OpenRouter reports them (spec 8, quotas)."""

import json
import urllib.request

from worker.remote.openrouter_endpoint import OPENROUTER_API


def openrouter_headroom(key: str, api: str = OPENROUTER_API, timeout: float = 10.0) -> int | None:
    """``free_model_daily_requests.remaining``; None when unknown, which means no headroom."""
    request = urllib.request.Request(f"{api}/key", headers={"Authorization": f"Bearer {key}"})
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            data = json.loads(response.read()).get("data") or {}
        return int(data["free_model_daily_requests"]["remaining"])
    except (OSError, ValueError, KeyError, TypeError):
        return None
