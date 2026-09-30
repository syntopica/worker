"""Dispatch a task to its runner's invocation."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.agy_invocation import agy_invocation
from worker.tasks.codex_invocation import codex_invocation
from worker.tasks.cursor_invocation import cursor_invocation
from worker.tasks.runner_invocation import RunnerInvocation


def task_invocation(
    profile: TaskProfile, prompt: str, workspace: Path, schema: Path | None, scratch: Path
) -> RunnerInvocation:
    """The profile's runner decides; the producer's ``runner`` was checked at submit."""
    if profile.runner == "codex":
        return codex_invocation(profile, prompt, workspace, schema, scratch)
    if profile.runner == "agy":
        return agy_invocation(profile, prompt, schema)
    return cursor_invocation(profile, prompt)
