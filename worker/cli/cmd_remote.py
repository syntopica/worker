"""`worker remote --name <node>`."""

import argparse
import sys
from pathlib import Path

from worker.cli.base_url import base_url
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.config_source import ConfigSource
from worker.config.load_worker_config import load_worker_config
from worker.node.coordinator_link import CoordinatorLink
from worker.remote.run_remote_node import run_remote_node


def cmd_remote(args: argparse.Namespace) -> int:
    """Run the OpenRouter loop with the node's own token and its local key file."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    node = config.nodes.get(args.name)
    if node is None:
        print(f"worker: node {args.name} is not in the configuration", file=sys.stderr)
        return 2
    if node.openrouter_key_file is None:
        print(f"worker: node {args.name} has no openrouter_key_file", file=sys.stderr)
        return 2
    try:
        key = Path(node.openrouter_key_file).expanduser().read_text().strip()
    except OSError:
        print("worker: the OpenRouter key file cannot be read", file=sys.stderr)
        return 2
    link = CoordinatorLink(base_url(config), read_token(state, args.name), args.name)
    run_remote_node(config, args.name, link, key, current=ConfigSource(config_path, config).current)
    return 0
