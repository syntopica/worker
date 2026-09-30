"""`worker quality` - how each model does, per queue and tier."""

import argparse
import json

from worker.cli.base_url import base_url
from worker.cli.fetch_status import fetch_status
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.load_worker_config import load_worker_config


def cmd_quality(args: argparse.Namespace) -> int:
    """Print attempt outcomes, then producer ratings, by queue, tier, provider and model."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    token = read_token(state, "admin")
    report = fetch_status(base_url(config), token, f"/v1/quality?days={args.days}")
    if args.json:
        print(json.dumps(report, indent=2))
        return 0
    for r in report["attempts"]:
        print(
            f"{r['queue']:22} {r['tier']:6} {r['provider']:10} {r['model']:32} "
            f"attempts={r['attempts']} ok={r['succeeded']} schema={r['schema_violations']} "
            f"failed={r['failed']} preempted={r['preempted']} wall={r['mean_wall_s'] or 0:.0f}s"
        )
    for r in report["ratings"]:
        print(
            f"{r['queue']:22} {r['tier']:6} {r['provider']:10} {r['model']:32} "
            f"results={r['results']} rated={r['rated']} good={r['good']} "
            f"edited={r['edited']} discarded={r['discarded']}"
        )
    return 0
