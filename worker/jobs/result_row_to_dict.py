"""Shape one results row (joined with its output body) as the API dict."""

import json
import sqlite3
from typing import Any


def result_row_to_dict(r: sqlite3.Row) -> dict[str, Any]:
    """``r`` carries the results columns plus ``output``, the raw output body."""
    return {
        "seq": r["seq"],
        "result_id": r["result_id"],
        "job_id": r["job_id"],
        "control": r["control"],
        "detail": json.loads(r["detail"]) if r["detail"] else None,
        "output": json.loads(r["output"]) if r["output"] else None,
        "executor": json.loads(r["executor"]) if r["executor"] else None,
        "usage": json.loads(r["usage"]) if r["usage"] else None,
    }
