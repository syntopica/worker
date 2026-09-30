"""The last balanced JSON object in a text, or None."""

import json


def last_json_object(text: str) -> str | None:
    """Scan from each ``{`` with the JSON decoder and keep the last object that parses."""
    decoder = json.JSONDecoder()
    last: str | None = None
    index = text.find("{")
    while index != -1:
        try:
            value, end = decoder.raw_decode(text, index)
        except ValueError:
            index = text.find("{", index + 1)
            continue
        if isinstance(value, dict):
            last = text[index:end]
        index = text.find("{", end)
    return last
