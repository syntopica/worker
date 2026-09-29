"""`worker submit`."""

import argparse
import json
from pathlib import Path

from worker.cli.build_client import build_client


def cmd_submit(args: argparse.Namespace) -> int:
    """Submit the JSON job file and print the response."""
    job = json.loads(Path(args.file).read_text())
    print(json.dumps(build_client(args.producer).submit(job)))
    return 0
