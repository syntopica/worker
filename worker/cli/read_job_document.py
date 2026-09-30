"""Read a job document from a file, or from stdin when the path is ``-``."""

import json
import sys
from pathlib import Path
from typing import Any


def read_job_document(path: str) -> dict[str, Any]:
    """The parsed job; ``-`` lets a script pipe a job it built in memory."""
    raw = sys.stdin.read() if path == "-" else Path(path).read_text()
    job: dict[str, Any] = json.loads(raw)
    return job
