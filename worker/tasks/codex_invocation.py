"""The read-only codex exec call (the posture of clips' triage and grade runners)."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def codex_invocation(
    profile: TaskProfile, prompt: str, workspace: Path, schema: Path | None, scratch: Path
) -> RunnerInvocation:
    """``-s read-only`` and no ``--search``: the workspace is all it may read.

    The prompt goes on stdin (``-``), never in argv: an argv prompt starting
    with ``-`` would be parsed as options that can lift the sandbox, and one
    equal to a subcommand name (``resume``, ``review``) would run it.

    ``--json`` prints events on stdout, whose ``turn.completed`` carries the
    usage; the answer still comes from the ``-o`` file.
    """
    answer = scratch / "last-message.txt"
    argv = [profile.command or "codex", "exec", "-", "-C", str(workspace), "-s", "read-only"]
    argv += ["--skip-git-repo-check", "--json", "-o", str(answer)]
    if profile.model:
        argv += ["-m", profile.model]
    if profile.reasoning:
        argv += ["-c", f"model_reasoning_effort={profile.reasoning}"]
    if schema is not None:
        argv += ["--output-schema", str(schema)]
    return RunnerInvocation(argv, prompt, answer)
