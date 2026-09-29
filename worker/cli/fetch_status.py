"""GET /v1/status as the admin principal."""

import json
import urllib.error
import urllib.request
from typing import Any

from worker.client.api_failure import ApiFailure
from worker.client.error_code import error_code


def fetch_status(base_url: str, token: str) -> dict[str, Any]:
    """The status payload; a refusal raises ApiFailure."""
    request = urllib.request.Request(
        f"{base_url}/v1/status", headers={"Authorization": f"Bearer {token}"}
    )
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            return dict(json.load(response))
    except urllib.error.HTTPError as error:
        raise ApiFailure(error.code, error_code(error.read())) from None
