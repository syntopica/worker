"""How one runner process is started: argv, stdin, and the file holding its answer."""

from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class RunnerInvocation:
    """``stdin`` None closes stdin at once; ``answer_file`` None means stdout."""

    argv: list[str]
    stdin: str | None
    answer_file: Path | None
