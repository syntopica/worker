"""Hold a reported error code to the fixed vocabulary."""

import re

from worker.jobs.error_codes import ERROR_CODES

_HTTP_CODE = re.compile(r"http_\d{3}")


def known_error_code(code: str | None) -> str | None:
    """The code itself when known, else ``executor_error``: free text never reaches storage."""
    if code is None or code in ERROR_CODES or _HTTP_CODE.fullmatch(code):
        return code
    return "executor_error"
