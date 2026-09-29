"""`worker jobs`."""

import argparse
import json

from worker.cli.build_client import build_client


def cmd_jobs(args: argparse.Namespace) -> int:
    """Print the producer's unacknowledged results for a queue as JSON."""
    print(json.dumps(build_client(args.producer).results(args.queue, 0, 100, 0), indent=2))
    return 0
