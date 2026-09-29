"""Turn an Ollama chat answer into the contract's output and usage."""

import json
from typing import Any


def chat_output(answer: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any]]:
    """``json`` is the parsed content when it is a JSON document, else None."""
    text = (answer.get("message") or {}).get("content") or ""
    try:
        parsed: Any = json.loads(text)
    except ValueError:
        parsed = None
    usage = {
        "tokens_in": answer.get("prompt_eval_count", 0),
        "tokens_out": answer.get("eval_count", 0),
    }
    return {"text": text, "json": parsed}, usage
