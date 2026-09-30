"""The read-only codex exec call (the posture of clips' triage and grade runners)."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def codex_invocation(
    profile: TaskProfile, prompt: str, workspace: Path, schema: Path | None, scratch: Path
) -> RunnerInvocation:
    """``-s read-only`` and no ``--search``: the workspace is all it may read."""
    answer = scratch / "last-message.txt"
    argv = [profile.command or "codex", "exec", prompt, "-C", str(workspace), "-s", "read-only"]
    argv += ["--skip-git-repo-check", "-o", str(answer)]
    if profile.model:
        argv += ["-m", profile.model]
    if profile.reasoning:
        argv += ["-c", f"model_reasoning_effort={profile.reasoning}"]
    if schema is not None:
        argv += ["--output-schema", str(schema)]
    return RunnerInvocation(argv, None, answer)
