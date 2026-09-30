"""`worker costs` - the usage ledger."""

import argparse
import json

from worker.cli.base_url import base_url
from worker.cli.fetch_status import fetch_status
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.load_worker_config import load_worker_config


def cmd_costs(args: argparse.Namespace) -> int:
    """Print attempts, tokens, cost and wall time by day, provider and queue."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    token = read_token(state, "admin")
    rows = fetch_status(base_url(config), token, f"/v1/costs?days={args.days}")["rows"]
    if args.json:
        print(json.dumps(rows, indent=2))
        return 0
    for r in rows:
        print(
            f"{r['day']} {r['provider']:11} {r['queue']:22} attempts={r['attempts']} "
            f"ok={r['succeeded']} in={r['tokens_in']} out={r['tokens_out']} "
            f"cost=${r['cost_usd']:.4f} wall={r['wall_s']:.0f}s"
        )
    return 0
