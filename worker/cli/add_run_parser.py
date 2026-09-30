"""The `worker run` subparser."""

import argparse

from worker.cli.cmd_run import cmd_run


def add_run_parser(sub: "argparse._SubParsersAction[argparse.ArgumentParser]") -> None:
    """Register `worker run --producer P [--wait S] [--poll S] FILE`."""
    run = sub.add_parser("run")
    run.add_argument("--producer", required=True)
    run.add_argument("--wait", type=float, default=3600.0)
    run.add_argument("--poll", type=float, default=5.0)
    run.add_argument("file")
    run.set_defaults(run=cmd_run)
