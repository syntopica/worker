"""Read a finished runner's answer from wherever its runner leaves it."""

from pathlib import Path

from worker.tasks.agy_answer import agy_answer
from worker.tasks.claude_answer import claude_answer
from worker.tasks.cursor_answer import cursor_answer
from worker.tasks.max_lane_answer import max_lane_answer


def read_answer(runner: str, answer_file: Path | None, stdout: str) -> str | None:
    """Codex writes a file; agy, cursor, max-lane and claude wrap it in a JSON envelope."""
    if answer_file is not None:
        try:
            text = answer_file.read_text()
        except OSError:
            return None
        return text if text.strip() else None
    if runner == "agy":
        return agy_answer(stdout)
    if runner == "max-lane":
        return max_lane_answer(stdout)
    if runner == "claude":
        return claude_answer(stdout)
    return cursor_answer(stdout)
