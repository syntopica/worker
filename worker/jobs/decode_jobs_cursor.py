"""Read back a cursor written by ``encode_jobs_cursor``."""

from worker.jobs.api_error import ApiError


def decode_jobs_cursor(cursor: str) -> tuple[float, str]:
    """``(created, id)``; anything else is ``400 bad_cursor``."""
    created, _, job_id = cursor.partition(":")
    try:
        when = float(created)
    except ValueError:
        raise ApiError(400, "bad_cursor") from None
    if not job_id:
        raise ApiError(400, "bad_cursor")
    return when, job_id
