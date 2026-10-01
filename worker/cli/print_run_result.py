"""Print one collected result for `worker run` and return its exit code."""

import json
import sys
from typing import Any

from worker.cli.format_executor import format_executor


def print_run_result(job_id: str, result: dict[str, Any]) -> int:
    """0 with the output on stdout and the result id on stderr; 1 with the control and its code on stderr.

    The JSON output wins over the text, the same choice every producer makes,
    so a caller with a schema reads validated JSON either way.
    """
    if result["control"] is None:
        # The result id lets the script rate the output once it has judged it,
        # and the executor lets it say which provider and model answered.
        executor = format_executor(result.get("executor"))
        print(
            f"worker job {job_id} result {result['result_id']} executor {executor}", file=sys.stderr
        )
        output = result["output"] or {}
        if output.get("json") is not None:
            print(json.dumps(output["json"]))
        else:
            print(output.get("text") or "")
        return 0
    code = (result.get("detail") or {}).get("error", "")
    print(f"worker job {job_id} ended {result['control']} {code}".rstrip(), file=sys.stderr)
    return 1
