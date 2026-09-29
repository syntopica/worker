"""Delay before a failed job is retried."""


def retry_backoff(attempts: int) -> float:
    """30 s doubling per failed attempt, capped at 15 minutes."""
    return float(min(30 * 2 ** (attempts - 1), 900))
