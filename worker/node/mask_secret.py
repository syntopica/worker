"""Replace a credential match's secret, or the whole match when it names none."""

import re

_MARK = "[REDACTED]"


def mask_secret(match: re.Match[str]) -> str:
    """Keep the label around a ``secret`` group so the text still reads naturally."""
    if "secret" not in match.re.groupindex or match.group("secret") is None:
        return _MARK
    start, end = match.span("secret")
    offset = match.start()
    whole = match.group(0)
    return whole[: start - offset] + _MARK + whole[end - offset :]
