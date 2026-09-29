"""`worker cancel`."""

import argparse
import json

from worker.cli.build_client import build_client


def cmd_cancel(args: argparse.Namespace) -> int:
    """Print the state the job ended in."""
    print(json.dumps({"state": build_client(args.producer).cancel(args.job_id)}))
    return 0
