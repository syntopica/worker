"""`worker run`: submit one job and wait for its output."""

import argparse
import sys
import time

from worker.cli.build_client import build_client
from worker.cli.print_run_result import print_run_result
from worker.cli.read_job_document import read_job_document
from worker.jobs.states import RELEASABLE

EXIT_PARKED = 3
EXIT_PENDING = 4


def cmd_run(args: argparse.Namespace) -> int:
    """Print the output and acknowledge it, for a shell script that wants one answer.

    Exit 3 when the job is parked behind its runner's quota wall and 4 when it
    is still pending at ``--wait``: both leave the job queued, and running the
    same job again collects it by its idempotency key. A split is declined,
    since a script submitting one job has no children to submit.
    """
    client = build_client(args.producer)
    job_id = client.submit(read_job_document(args.file))["id"]
    deadline = time.monotonic() + args.wait
    while True:
        job = client.get(job_id)
        result = job["result"]
        if result is not None:
            split = result["control"] == "split_requested"
            client.ack(job_id, result["result_id"], decline=split)
            if not split:
                return print_run_result(job_id, result)
        elif job["state"] == "succeeded" or job["state"] in RELEASABLE:
            print(f"worker job {job_id} is {job['state']} with nothing to collect", file=sys.stderr)
            return 1
        if job.get("cooling_until") is not None:
            print(f"worker job {job_id} is parked behind a quota wall", file=sys.stderr)
            return EXIT_PARKED
        if time.monotonic() >= deadline:
            print(f"worker job {job_id} still {job['state']} after {args.wait:g}s", file=sys.stderr)
            return EXIT_PENDING
        time.sleep(args.poll)
