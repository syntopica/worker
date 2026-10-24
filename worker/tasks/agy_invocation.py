"""The agy print-mode call in its no-tools posture (clips' triage runner)."""

import json
from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.agy_no_tools_preface import AGY_NO_TOOLS_PREFACE
from worker.tasks.runner_invocation import RunnerInvocation


def agy_invocation(profile: TaskProfile, prompt: str, schema: Path | None) -> RunnerInvocation:
    """``--sandbox --mode plan --disable-slash-commands`` and never skip-permissions.

    A tool request therefore blocks rather than being approved; a task needs
    none because everything it may consider is in the prompt.

    The prompt is never an argument: argv is visible to every local process and
    agy pre-scans it for ``--help`` and ``--version``. It travels as the one
    ``user`` event of ``stream-json`` input on stdin (agy 1.4.6), JSON-escaped
    onto a single line. The prompt opens with a line asking for a plain answer,
    because a denied tool request ends the turn with an empty response.
    """
    minutes = max(1, int(profile.timeout_s // 60))
    argv = [profile.command or "agy", "--input-format", "stream-json"]
    argv += ["--output-format", "stream-json"]
    argv += ["--print-timeout", f"{minutes}m", "--sandbox", "--mode", "plan"]
    argv += ["--disable-slash-commands"]
    if profile.model:
        argv += ["--model", profile.model]
    if schema is not None:
        argv += ["--json-schema", str(schema)]
    content = f"{AGY_NO_TOOLS_PREFACE}{prompt}"
    event = {"event": "user", "message": {"content": content}}
    return RunnerInvocation(argv, json.dumps(event, ensure_ascii=True) + "\n", None)
