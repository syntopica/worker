"""A refusal from the coordinator, by status and code."""


class ApiFailure(Exception):  # noqa: N818 - public name fixed by contract v1
    """Raised by WorkerClient for any non-2xx answer."""

    def __init__(self, status: int, code: str) -> None:
        super().__init__(f"{status} {code}")
        self.status = status
        self.code = code
