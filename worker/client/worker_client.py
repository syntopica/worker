"""Producer client for contract v1 (stdlib only, so any project can vendor it)."""

import json
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

from worker.client.api_failure import ApiFailure
from worker.client.error_code import error_code
from worker.client.quote_job_id import quote_job_id


class WorkerClient:
    """Record ``result_id`` with your domain writes, then ``ack`` (spec 6).

    Coordinator refusals raise ``ApiFailure``. Transport errors (``URLError``,
    ``TimeoutError``) propagate unwrapped.
    """

    def __init__(self, base_url: str, token: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token

    def _call(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None = None,
        timeout: float = 40.0,
    ) -> Any:
        data = json.dumps(payload).encode() if payload is not None else None
        headers = {"Authorization": f"Bearer {self.token}", "Content-Type": "application/json"}
        request = urllib.request.Request(self.base_url + path, data, headers, method=method)
        try:
            with urllib.request.urlopen(request, timeout=timeout) as response:
                raw = response.read()
                return json.loads(raw) if raw else None
        except urllib.error.HTTPError as error:
            raise ApiFailure(error.code, error_code(error.read())) from None

    def submit(self, job: dict[str, Any]) -> dict[str, Any]:
        """``{"id", "created"}``."""
        return dict(self._call("POST", "/v1/jobs", job))

    def get(self, job_id: str) -> dict[str, Any]:
        """``{"id", "state", "error", "result"}``."""
        return dict(self._call("GET", f"/v1/jobs/{quote_job_id(job_id)}"))

    def results(
        self, queue: str, after: int = 0, limit: int = 50, wait: float = 0
    ) -> list[dict[str, Any]]:
        """Unacknowledged results after ``after``, long-polling up to ``wait`` seconds."""
        query = urllib.parse.urlencode(
            {"queue": queue, "after": after, "limit": limit, "wait": wait}
        )
        return list(self._call("GET", f"/v1/results?{query}")["results"])

    def ack(self, job_id: str, result_id: str, decline: bool = False) -> None:
        """Acknowledge, or decline a ``split_requested`` control result."""
        payload = {"result_id": result_id, "decline": decline}
        self._call("POST", f"/v1/jobs/{quote_job_id(job_id)}/ack", payload)

    def cancel(self, job_id: str) -> str:
        """The state the job ended in."""
        return str(self._call("POST", f"/v1/jobs/{quote_job_id(job_id)}/cancel", {})["state"])

    def wait_for(self, job_id: str, timeout: float, poll: float = 5.0) -> dict[str, Any] | None:
        """The job's latest unacknowledged result, control results included.

        None when ``timeout`` passes, which is also the answer for a job with no
        unacknowledged result, such as a cancelled or already acknowledged one.
        """
        deadline = time.monotonic() + timeout
        while True:
            result = self.get(job_id)["result"]
            if result is not None or time.monotonic() >= deadline:
                return dict(result) if result is not None else None
            time.sleep(min(poll, max(0.0, deadline - time.monotonic())))
