"""`worker audit` - sensitive reveals and admin actions."""

import argparse
import json
from datetime import UTC, datetime

from worker.cli.base_url import base_url
from worker.cli.fetch_status import fetch_status
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.load_worker_config import load_worker_config


def cmd_audit(args: argparse.Namespace) -> int:
    """Print the audit rows of the last ``--days`` days, newest first; never content."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    token = read_token(state, "admin")
    rows = fetch_status(base_url(config), token, f"/v1/admin/audit?days={args.days}")["rows"]
    if args.json:
        print(json.dumps(rows, indent=2))
        return 0
    for r in rows:
        when = datetime.fromtimestamp(r["time"], UTC).strftime("%Y-%m-%d %H:%M:%S")
        print(f"{when} {r['action']:6} {r['class']:8} {r['job_id']} {r['principal']}")
    return 0
