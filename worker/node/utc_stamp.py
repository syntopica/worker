"""The current UTC time as a second-precision ISO 8601 stamp."""

from datetime import UTC, datetime


def utc_stamp() -> str:
    """``YYYY-MM-DDTHH:MM:SSZ``, for log lines read back by time."""
    return datetime.now(UTC).strftime("%Y-%m-%dT%H:%M:%SZ")
