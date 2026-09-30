"""`worker status` and `worker nodes`."""

import argparse
import json

from worker.cli.base_url import base_url
from worker.cli.fetch_status import fetch_status
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.load_worker_config import load_worker_config


def cmd_status(args: argparse.Namespace) -> int:
    """Print queues (or nodes with `nodes`) as a table, or everything as JSON."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    status = fetch_status(base_url(config), read_token(state, "admin"))
    if args.json:
        print(json.dumps(status, indent=2))
    elif args.nodes:
        for name, node in status.get("nodes", {}).items():
            reason = node.get("reason") or "working/ready"
            last = node.get("last_release")
            released = f" last_release={last['code']}@{last['age_s']:.0f}s" if last else ""
            if "loads_1h" in node:
                released += f" loads_1h={node['loads_1h']} unloads_1h={node['unloads_1h']}"
            print(
                f"{name:20} {reason:22} idle={node.get('idle_s')} "
                f"resident={node.get('resident')} age={node.get('age_s', 0):.0f}s{released}"
            )
    else:
        for name, q in status.get("queues", {}).items():
            print(
                f"{name:24} {json.dumps(q.get('states'))} done_1h={q.get('done_1h')} "
                f"wasted_1h={q.get('wasted_1h_s', 0):.0f}s oldest={q.get('oldest_queued_s')}"
            )
        for runner, left in status.get("cooldowns", {}).items():
            print(f"cooldown {runner:15} {left:.0f}s left")
    return 0
