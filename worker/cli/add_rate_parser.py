"""The `worker rate` subparser."""

import argparse

from worker.cli.cmd_rate import cmd_rate
from worker.jobs.states import RATINGS


def add_rate_parser(sub: "argparse._SubParsersAction[argparse.ArgumentParser]") -> None:
    """Register `worker rate --producer P JOB_ID RESULT_ID RATING`."""
    rate = sub.add_parser("rate")
    rate.add_argument("--producer", required=True)
    rate.add_argument("job_id")
    rate.add_argument("result_id")
    rate.add_argument("rating", choices=RATINGS)
    rate.set_defaults(run=cmd_rate)
