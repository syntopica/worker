"""The Claude Code CLI print-mode call: no tools, no settings, no session file."""

from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.runner_invocation import RunnerInvocation


def claude_invocation(profile: TaskProfile, prompt: str, schema: Path | None) -> RunnerInvocation:
    """The prompt goes on stdin; the schema is passed inline, as the CLI expects.

    No settings means no hooks and no plugins, and no session file means the run
    never lands in a conversation archive. The account is the one the profile's
    ``env`` selects.
    """
    argv = [profile.command or "claude", "-p", "--model", profile.model or ""]
    argv += ["--output-format", "json", "--tools", "", "--strict-mcp-config"]
    argv += ["--setting-sources", "", "--no-session-persistence"]
    if schema is not None:
        argv += ["--json-schema", schema.read_text()]
    return RunnerInvocation(argv, prompt, None)
