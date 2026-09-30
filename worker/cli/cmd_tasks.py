"""`worker tasks --name <node>`."""

import argparse
import sys

from worker.cli.base_url import base_url
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.config_source import ConfigSource
from worker.config.load_worker_config import load_worker_config
from worker.node.coordinator_link import CoordinatorLink
from worker.tasks.run_task_node import run_task_node


def cmd_tasks(args: argparse.Namespace) -> int:
    """Run the read-only task loop with the node's own token."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    if args.name not in config.nodes:
        print(f"worker: node {args.name} is not in the configuration", file=sys.stderr)
        return 2
    link = CoordinatorLink(base_url(config), read_token(state, args.name), args.name)
    run_task_node(config, args.name, link, current=ConfigSource(config_path, config).current)
    return 0
