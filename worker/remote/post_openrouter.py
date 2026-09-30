"""POST one chat completion to OpenRouter."""

import json
import urllib.error
import urllib.request
from typing import Any

from worker.remote.openrouter_endpoint import OPENROUTER_API


def post_openrouter(
    key: str, body: dict[str, Any], timeout: float, api: str = OPENROUTER_API
) -> tuple[dict[str, Any] | None, str | None]:
    """(answer, None) on success, (None, code) otherwise; codes are allowlisted.

    ``http_<status>`` for a refusal, ``transport_error`` when unreachable,
    ``bad_response`` for a body that is not a completion.
    """
    request = urllib.request.Request(
        f"{api}/chat/completions",
        json.dumps(body).encode(),
        {"Authorization": f"Bearer {key}", "Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            answer = json.loads(response.read())
    except urllib.error.HTTPError as error:
        return None, f"http_{error.code}"
    except (urllib.error.URLError, TimeoutError, OSError):
        return None, "transport_error"
    except ValueError:
        return None, "bad_response"
    if not isinstance(answer, dict) or "error" in answer or not answer.get("choices"):
        return None, "bad_response"
    return answer, None
