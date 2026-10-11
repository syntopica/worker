"""The text a runner's failure is named from: its own messages, never the model's."""

from worker.tasks.codex_failure_text import codex_failure_text


def runner_failure_text(runner: str, stdout: str, stderr: str) -> str:
    """Stderr, followed for codex by the error events of its ``--json`` stream."""
    if runner == "codex":
        return stderr + "\n" + codex_failure_text(stdout)
    return stderr
