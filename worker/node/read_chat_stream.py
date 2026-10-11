"""Read a streamed /api/chat response into one answer."""

import json
from collections.abc import Callable
from typing import Any


def read_chat_stream(
    response: Any, on_first_chunk: Callable[[], None]
) -> tuple[dict[str, Any] | None, str | None]:
    """(answer, None) once a ``done`` chunk arrives, else (None, ``bad_response``).

    The answer is the final chunk, which carries the counts, with the content
    of every chunk joined into its message. ``on_first_chunk`` runs when the
    first chunk arrives: prefill is over and generation has started. A chunk
    that is not a JSON object or carries ``error`` fails the call.
    """
    parts: list[str] = []
    for raw in iter(response.readline, b""):
        if not raw.strip():
            continue
        if not parts:
            on_first_chunk()
        try:
            chunk = json.loads(raw)
        except ValueError:
            return None, "bad_response"
        if not isinstance(chunk, dict) or "error" in chunk:
            return None, "bad_response"
        message = chunk.get("message")
        content = message.get("content") if isinstance(message, dict) else None
        parts.append(content if isinstance(content, str) else "")
        if chunk.get("done") is True:
            final = message if isinstance(message, dict) else {}
            return {**chunk, "message": {**final, "content": "".join(parts)}}, None
    return None, "bad_response"
