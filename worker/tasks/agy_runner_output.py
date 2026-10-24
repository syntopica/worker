"""The agy text a quota wall is judged from: the runner's own, never the model's."""

from worker.tasks.agy_final_result import agy_final_result


def agy_runner_output(stdout: str, stderr: str, answered: bool) -> str:
    """Stderr plus, for a failed final result, its explicit error field.

    A successful result and every step update are model output over untrusted
    input and may quote a wall sentence, so a stream contributes only the error
    of a failed final result. Legacy envelopes contribute only their explicit error field.
    """
    final = agy_final_result(stdout)
    if answered or final is None:
        return stderr
    result, streamed = final
    if streamed and result.get("status") == "SUCCESS":
        return stderr
    error = result.get("error")
    return stderr + ("\n" + error if isinstance(error, str) else "")
