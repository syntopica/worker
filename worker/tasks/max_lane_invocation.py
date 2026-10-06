"""The ``max-lane-run`` call: one forced-tool request, the prompt on stdin."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def max_lane_invocation(profile: TaskProfile, prompt: str, schema: Path | None) -> RunnerInvocation:
    """No tools and no workspace: everything the model may consider is in the prompt."""
    argv = [profile.command or "max-lane-run", "--model", profile.model or ""]
    if schema is not None:
        argv += ["--schema", str(schema)]
    return RunnerInvocation(argv, prompt, None)
