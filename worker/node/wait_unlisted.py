"""Poll /api/ps until a model is no longer listed, heartbeating between polls."""

from collections.abc import Callable

_POLLS = 5
_POLL_S = 2.0


def wait_unlisted(
    url: str,
    model: str,
    heartbeat: Callable[[], bool],
    resident: Callable[[str], list[str] | None],
    pause: Callable[[float], None],
) -> list[str] | None:
    """The last /api/ps answer: without ``model`` on success, None if it never answered."""
    listed: list[str] | None = None
    for _ in range(_POLLS):
        heartbeat()
        listed = resident(url)
        if listed is not None and model not in listed:
            return listed
        pause(_POLL_S)
    return listed
