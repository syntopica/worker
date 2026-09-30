"""Whether a finished runner hit its account's quota rather than failing on the task."""

from worker.tasks.quota_wall_patterns import QUOTA_WALL_PATTERNS


def is_quota_wall(runner: str, exit_code: int, output: str, answer: str | None) -> bool:
    """The runner's wall sentence anywhere in its output, or agy's silent wall.

    agy exits 0 with no answer when its quota is spent; that shape is read as a
    wall too, since nothing else in a no-tools run produces it.
    """
    pattern = QUOTA_WALL_PATTERNS.get(runner)
    if pattern is not None and pattern.search(output):
        return True
    return runner == "agy" and exit_code == 0 and answer is None
