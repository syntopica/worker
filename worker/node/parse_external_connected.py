"""Whether the machine is on AC power, from the battery's ``ExternalConnected`` flag."""

import re

_FLAG = re.compile(r'"ExternalConnected"\s*=\s*(Yes|No)')


def parse_external_connected(text: str) -> bool | None:
    """None when ``ioreg`` lists no battery (a desktop) or the flag is absent."""
    match = _FLAG.search(text)
    return None if match is None else match.group(1) == "Yes"
