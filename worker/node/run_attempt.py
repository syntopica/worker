"""Execute one leased inference job, releasing it the moment the host says so."""

from collections.abc import Callable
from typing import Any

from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.node.attempt_report import attempt_report
from worker.node.chat_output import chat_output
from worker.node.drain_backend import drain_backend
from worker.node.host_state import HostState
from worker.node.ollama_call import OllamaCall
from worker.node.ollama_request_body import ollama_request_body
from worker.node.release_reason import release_reason

_SAMPLE_S = 2.0
_CALL_TIMEOUT_S = 1800.0


def run_attempt(  # noqa: PLR0913, PLR0917
    lease: dict[str, Any],
    link: Any,
    pin: ModelPin,
    node: NodePolicy,
    sample: Callable[[], HostState],
    sleep: Callable[[float], None],
    clock: Callable[[], float],
) -> str:
    """Return the outcome; completion is reported unless the attempt was fenced out."""
    attempt, gen = lease["attempt_id"], lease["generation"]
    started = clock()
    executor = {"node": node.name, "provider": "ollama", "model": pin.name}
    call = OllamaCall.start(
        node.ollama_url, ollama_request_body(pin, lease["input"]), _CALL_TIMEOUT_S
    )
    while not call.done():
        sleep(_SAMPLE_S)
        if call.done():
            break
        if not link.heartbeat(attempt, gen, False):
            call.cancel()
            quiet = drain_backend(node.ollama_url, pin, node.ollama_launchd_label, lambda: True)
            return "fenced" if quiet else "drain_failed"
        if call.done():
            break
        reason = release_reason(sample(), lease["run_when"])
        if reason is not None and not call.done():
            call.cancel()
            quiet = drain_backend(
                node.ollama_url,
                pin,
                node.ollama_launchd_label,
                lambda: link.heartbeat(attempt, gen, True),
            )
            report = attempt_report("preempted", executor, clock() - started, error_code=reason)
            link.complete(attempt, gen, report)
            return "preempted" if quiet else "drain_failed"
    answer, error = call.outcome()
    if answer is None:
        report = attempt_report("failed", executor, clock() - started, error_code=error)
        link.complete(attempt, gen, report)
        return "failed"
    output, usage = chat_output(answer)
    report = attempt_report("succeeded", executor, clock() - started, output=output, usage=usage)
    link.complete(attempt, gen, report)
    return "succeeded"
