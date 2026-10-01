"""Turn a chat-completions answer into the contract's output and usage."""

from typing import Any

from worker.node.parse_json_answer import parse_json_answer


def openrouter_output(answer: dict[str, Any]) -> tuple[dict[str, Any], dict[str, Any]]:
    """``json`` is the parsed content when it is a JSON document, else None.

    A single surrounding code fence is tolerated: some free models add one
    even under a JSON response format.
    """
    choices = answer.get("choices") or [{}]
    text = str(((choices[0] or {}).get("message") or {}).get("content") or "")
    parsed = parse_json_answer(text)
    usage = answer.get("usage") or {}
    return {"text": text, "json": parsed}, {
        "tokens_in": usage.get("prompt_tokens", 0),
        "tokens_out": usage.get("completion_tokens", 0),
        "cost_usd": float(usage.get("cost") or 0.0),
    }
