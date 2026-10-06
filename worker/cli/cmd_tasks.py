"""`worker tasks --name <node>`."""

import argparse
import sys

from worker.cli.base_url import base_url
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.config.config_source import ConfigSource
from worker.config.load_worker_config import load_worker_config
from worker.node.coordinator_link import CoordinatorLink
from worker.node.raise_on_sigterm import raise_on_sigterm
from worker.tasks.run_on_demand import run_on_demand
from worker.tasks.run_task_slots import run_task_slots


def cmd_tasks(args: argparse.Namespace) -> int:
    """Run the read-only task loop with the node's own token.

    With ``--profile`` it drains that on-demand profile's jobs instead and
    exits once none is left; ``--slots`` overrides the node's task slots.
    """
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    if args.name not in config.nodes:
        print(f"worker: node {args.name} is not in the configuration", file=sys.stderr)
        return 2
    token = read_token(state, args.name)
    slots = args.slots or config.nodes[args.name].task_slots
    links = [CoordinatorLink(base_url(config), token, args.name) for _ in range(slots)]
    raise_on_sigterm()
    if args.profile is not None:
        ran = run_on_demand(config, args.name, links, args.profile)
        print(f"worker: on-demand {args.profile} ran {ran} jobs")
        return 0
    run_task_slots(config, args.name, links, current=ConfigSource(config_path, config).current)
    return 0
