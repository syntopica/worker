"""The whole wall sentences each runner prints (clips measured codex and agy).

Whole sentences, never the memorable half: stdout carries the model's own
prose over untrusted input, and a text saying "I ran out of credits" must not
rest an account.
"""

import re

QUOTA_WALL_PATTERNS = {
    "codex": re.compile(
        r"your workspace is out of credits\.\s*ask your workspace owner to refill", re.I
    ),
    "agy": re.compile(r"individual quota reached\.\s*please upgrade your subscription", re.I),
    # Both measured by atrium: the monthly window on a Cursor model
    # (2026-09-30) and a third-party model's window (2026-09-16).
    "cursor": re.compile(
        r"you['\u2019]re out of usage\.\s*switch to auto, or ask your admin to increase your limit"
        r"|you['\u2019]ve hit your usage limit[^\n]{0,300}switch to a different model or set a"
        r"\s+spend limit",
        re.I,
    ),
    # max-lane-run's own stderr once every Keychain account answered 429.
    "max-lane": re.compile(
        r"max-lane-run: max lane request failed for keychain service [^\n]{1,80}"
        r" with http status 429\.",
        re.I,
    ),
}
