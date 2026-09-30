"""The agy print-mode call in its no-tools posture (clips' triage runner)."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def agy_invocation(profile: TaskProfile, prompt: str, schema: Path | None) -> RunnerInvocation:
    """``--sandbox --mode plan --disable-slash-commands`` and never skip-permissions.

    A tool request therefore blocks rather than being approved; a task needs
    none because everything it may consider is in the prompt.
    """
    minutes = max(1, int(profile.timeout_s // 60))
    argv = [profile.command or "agy", "-p", prompt, "--output-format", "json"]
    argv += ["--print-timeout", f"{minutes}m", "--sandbox", "--mode", "plan"]
    argv += ["--disable-slash-commands"]
    if profile.model:
        argv += ["--model", profile.model]
    if schema is not None:
        argv += ["--json-schema", str(schema)]
    return RunnerInvocation(argv, None, None)
