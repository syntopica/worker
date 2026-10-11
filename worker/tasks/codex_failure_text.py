"""The error messages codex itself prints as ``--json`` events."""

from worker.tasks.jsonl_event import jsonl_event


def codex_failure_text(stdout: str) -> str:
    """The ``error`` and ``turn.failed`` messages, one per line.

    With ``--json`` codex moves its own failures from stderr to these stdout
    events. Agent messages and ``item.completed`` items of type ``error``
    (configuration warnings) are left out, so the model's prose over untrusted
    input never names a failure.
    """
    messages = []
    for line in stdout.splitlines():
        event = jsonl_event(line)
        if event is None:
            continue
        if event.get("type") == "error":
            message = event.get("message")
        elif event.get("type") == "turn.failed":
            error = event.get("error")
            message = error.get("message") if isinstance(error, dict) else None
        else:
            continue
        if isinstance(message, str):
            messages.append(message)
    return "\n".join(messages)
