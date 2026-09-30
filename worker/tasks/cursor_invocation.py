"""The cursor-agent print-mode call in ask mode, prompt on stdin (clips' grade runner)."""

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def cursor_invocation(profile: TaskProfile, prompt: str) -> RunnerInvocation:
    """``--mode ask`` keeps tools away; ``--trust`` grants only the workspace.

    The prompt goes on stdin: an argv prompt of a few hundred KB makes the CLI
    exit 0 with no output at all (measured by clips, 2026-09-11).
    """
    argv = [profile.command or "cursor-agent", "-p", "--trust", "--mode", "ask"]
    argv += ["--output-format", "json"]
    if profile.model:
        argv += ["--model", profile.model]
    return RunnerInvocation(argv, prompt, None)
