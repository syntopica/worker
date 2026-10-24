"""Whether a finished runner hit its account's quota rather than failing on the task."""

from worker.tasks.quota_wall_patterns import QUOTA_WALL_PATTERNS


def is_quota_wall(
    runner: str, exit_code: int, output: str, answer: str | None, *, silent_wall: bool = True
) -> bool:
    """The runner's wall sentence anywhere in its output, or agy's silent wall.

    agy exits 0 with no answer when its quota is spent; that shape is read as a
    wall too, only for legacy envelopes; an incomplete stream is not quota evidence. A cursor run
    or max-lane run that answered is never a wall: its stdout carries the
    model's answer, which may quote the wall sentence from the material it was
    given.
    """
    if runner in {"cursor", "max-lane"} and answer is not None:
        return False
    pattern = QUOTA_WALL_PATTERNS.get(runner)
    if pattern is not None and pattern.search(output):
        return True
    return silent_wall and runner == "agy" and exit_code == 0 and answer is None
