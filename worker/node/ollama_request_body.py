"""Build an /api/chat body with the model's pinned options (spec 5)."""

from typing import Any

from worker.config.model_pin import ModelPin
from worker.node.sampling_keys import SAMPLING_KEYS


def ollama_request_body(pin: ModelPin, job_input: dict[str, Any]) -> dict[str, Any]:
    """``num_ctx`` and ``keep_alive`` always come from the pin, never the job."""
    options = {k: v for k, v in (job_input.get("options") or {}).items() if k in SAMPLING_KEYS}
    body: dict[str, Any] = {
        "model": pin.name,
        "messages": job_input.get("messages") or [],
        "stream": False,
        "think": False,
        "keep_alive": pin.keep_alive,
        "options": {**options, "num_ctx": pin.num_ctx},
    }
    if job_input.get("schema"):
        body["format"] = job_input["schema"]
    elif job_input.get("format") == "json":
        body["format"] = "json"
    return body
