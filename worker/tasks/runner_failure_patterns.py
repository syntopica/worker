"""Signs in a runner's own messages that tell why it exited without an answer.

Read only on the tail of the runner's own messages (stderr, and codex's error
events), never the model's prose: the code is a label for the owner, it rests
nothing, and the text itself is never kept.
"""

import re

RUNNER_FAILURE_PATTERNS = (
    (
        "runner_auth",
        re.compile(
            r"\b401\b|unauthori[sz]ed|not logged in|please (?:log|sign) in"
            r"|authentication (?:failed|required)|invalid api key",
            re.I,
        ),
    ),
    ("rate_limited", re.compile(r"\b429\b|rate.?limit|too many requests", re.I)),
    (
        "runner_unavailable",
        re.compile(
            r"\b50[234]\b|overloaded|service unavailable|temporarily unavailable"
            r"|stream disconnected|connection (?:reset|refused|closed)|timed out",
            re.I,
        ),
    ),
)
