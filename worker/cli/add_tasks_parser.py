"""The `worker tasks` subparser."""

import argparse

from worker.cli.cmd_tasks import cmd_tasks


def add_tasks_parser(sub: "argparse._SubParsersAction[argparse.ArgumentParser]") -> None:
    """Register `worker tasks --name N [--profile P] [--slots K]`."""
    tasks = sub.add_parser("tasks")
    tasks.add_argument("--name", required=True)
    tasks.add_argument("--profile")
    tasks.add_argument("--slots", type=int)
    tasks.set_defaults(run=cmd_tasks)
