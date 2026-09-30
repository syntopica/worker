"""The whole wall sentences each runner prints (clips measured both).

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
}
