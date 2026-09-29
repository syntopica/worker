"""`worker node --name <node>`."""

import argparse

from worker.cli.base_url import base_url
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.load_worker_config import load_worker_config
from worker.node.coordinator_link import CoordinatorLink
from worker.node.run_node import run_node


def cmd_node(args: argparse.Namespace) -> int:
    """Run the node loop against the configured coordinator address."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    link = CoordinatorLink(base_url(config), read_token(state, args.name), args.name)
    run_node(config, args.name, link)
    return 0
