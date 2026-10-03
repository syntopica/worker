"""Run one leased task end to end and return its completion report."""

import json
import shutil
import tempfile
from collections.abc import Callable
from pathlib import Path
from typing import Any

from worker.config.task_profile import TaskProfile
from worker.jobs.known_error_code import known_error_code
from worker.node.attempt_report import attempt_report
from worker.node.redact_credentials import redact_credentials
from worker.tasks.build_workspace import build_workspace
from worker.tasks.judge_run import judge_run
from worker.tasks.read_answer import read_answer
from worker.tasks.runner_failure_code import runner_failure_code
from worker.tasks.start_runner import start_runner
from worker.tasks.supervise_process import supervise_process
from worker.tasks.task_invocation import task_invocation
from worker.tasks.workspace_error import WorkspaceError


def run_task(  # noqa: PLR0913, PLR0917
    profile: TaskProfile,
    lease: dict[str, Any],
    node: str,
    heartbeat: Callable[[], bool],
    clock: Callable[[], float],
    sleep: Callable[[float], None],
) -> dict[str, Any] | None:
    """The report to complete with, or None when the attempt was fenced out.

    The scratch directory (workspace, schema, streams) is removed however the
    run ends, so no task input outlives its attempt on disk. Every runner is a
    remote model, so credentials are redacted from the prompt first; input
    files are passed as they are.
    """
    task = lease["input"]
    executor = {"node": node, "provider": profile.runner, "model": profile.model or ""}
    started = clock()
    scratch = Path(tempfile.mkdtemp(prefix="worker-task-"))
    try:
        workspace = scratch / "workspace"
        workspace.mkdir()
        try:
            inputs = list(task.get("inputs") or [])
            build_workspace(profile.input_root, inputs, workspace, profile.denied_root)
        except WorkspaceError as error:
            return attempt_report("failed", executor, 0.0, error_code=known_error_code(str(error)))
        schema = None
        if task.get("output_schema") is not None:
            schema = scratch / "output-schema.json"
            schema.write_text(json.dumps(task["output_schema"]))
        prompt = redact_credentials(str(task["prompt"]))
        invocation = task_invocation(profile, prompt, workspace, schema, scratch)
        with (scratch / "stdout").open("wb") as out, (scratch / "stderr").open("wb") as err:
            try:
                proc = start_runner(profile, invocation, workspace, (out, err))
            except OSError:
                return attempt_report("failed", executor, 0.0, error_code="runner_failed")
            ending = supervise_process(proc, profile.timeout_s, heartbeat, clock, sleep)
        wall_s = clock() - started
        if ending == "fenced":
            return None
        if ending == "timeout":
            return attempt_report("failed", executor, wall_s, error_code="timeout")
        stdout = (scratch / "stdout").read_text(errors="replace")
        stderr = (scratch / "stderr").read_text(errors="replace")
        answer = read_answer(profile.runner, invocation.answer_file, stdout)
        outcome, code, output = judge_run(
            profile.runner, proc.returncode or 0, stdout + stderr, answer
        )
        if code == "runner_failed":
            code = runner_failure_code(stderr)
        return attempt_report(outcome, executor, wall_s, output=output, error_code=code)
    finally:
        shutil.rmtree(scratch, ignore_errors=True)
