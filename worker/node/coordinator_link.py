"""The node's side of the API (Task 9 routes)."""

import http.client
import json
import urllib.error
import urllib.request
from http import HTTPStatus
from typing import Any

_FENCED = 409


class CoordinatorLink:
    """A 409 on heartbeat means the attempt was fenced out: stop, do not complete.

    Transport errors never raise: the coordinator may be restarting, and lease
    expiry is the safety net for anything lost meanwhile.
    """

    def __init__(self, base_url: str, token: str, node: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.node = node

    def _post(self, path: str, payload: dict[str, Any]) -> tuple[int | None, dict[str, Any] | None]:
        """Status and body; status None means the coordinator could not be reached."""
        request = urllib.request.Request(
            self.base_url + path,
            json.dumps(payload).encode(),
            {"Authorization": f"Bearer {self.token}", "Content-Type": "application/json"},
        )
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                raw = response.read()
                return response.status, (json.loads(raw) if raw else None)
        except urllib.error.HTTPError as error:
            return error.code, None
        except (urllib.error.URLError, http.client.HTTPException, OSError, ValueError):
            return None, None

    def lease(
        self, resident: str | None, user_active: bool, free_gb: float, idle_s: float
    ) -> dict[str, Any] | None:
        """A lease dict, or None for 204 / refusal / unreachable coordinator."""
        status, body = self._post(
            "/v1/leases",
            {
                "node": self.node,
                "resident_model": resident,
                "user_active": user_active,
                "free_gb": free_gb,
                "current_idle_s": idle_s,
            },
        )
        return body if status == HTTPStatus.OK else None

    def lease_task(self, user_active: bool, idle_s: float) -> dict[str, Any] | None:
        """A task lease for the ``worker tasks`` loop, or None."""
        status, body = self._post(
            "/v1/leases",
            {
                "node": self.node,
                "kind": "task",
                "user_active": user_active,
                "current_idle_s": idle_s,
            },
        )
        return body if status == HTTPStatus.OK else None

    def lease_remote(self) -> dict[str, Any] | None:
        """An escalated inference lease for the ``worker remote`` loop, or None."""
        status, body = self._post(
            "/v1/leases",
            {"node": self.node, "kind": "openrouter", "user_active": False, "current_idle_s": 0},
        )
        return body if status == HTTPStatus.OK else None

    def heartbeat(self, attempt_id: str, generation: int, draining: bool) -> bool:
        """False only when fenced out (409); an unreachable coordinator keeps the attempt."""
        status, _ = self._post(
            f"/v1/attempts/{attempt_id}/heartbeat",
            {"generation": generation, "draining": draining},
        )
        return status != _FENCED

    def complete(self, attempt_id: str, generation: int, report: dict[str, Any]) -> None:
        """Best effort; a lost completion is reclaimed by lease expiry."""
        self._post(f"/v1/attempts/{attempt_id}/complete", {**report, "generation": generation})

    def report(self, state: dict[str, Any]) -> None:
        """Metadata-only host report for `worker status`."""
        self._post(f"/v1/nodes/{self.node}/report", state)
