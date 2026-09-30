"""Turn a finished runner into an outcome, an error code and an output."""

from typing import Any

from worker.tasks.is_quota_wall import is_quota_wall
from worker.tasks.task_output import task_output


def judge_run(
    runner: str, exit_code: int, output: str, answer: str | None
) -> tuple[str, str | None, dict[str, Any] | None]:
    """A wall outranks everything; an answer outranks a non-zero exit.

    codex has exited non-zero after writing a usable last message, and the
    coordinator's schema check decides whether that answer is acceptable.
    """
    if is_quota_wall(runner, exit_code, output, answer):
        return "failed", "quota_wall", None
    if answer is None:
        return "failed", "no_output" if exit_code == 0 else "runner_failed", None
    return "succeeded", None, task_output(answer)
