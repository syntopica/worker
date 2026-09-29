"""An error the API returns as ``{"error": code}`` with an HTTP status."""


class ApiError(Exception):
    """Carries only an allowlisted code, never request or provider content."""

    def __init__(self, status: int, code: str) -> None:
        super().__init__(f"{status} {code}")
        self.status = status
        self.code = code
