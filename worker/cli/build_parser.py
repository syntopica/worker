"""The `worker` argument parser."""

import argparse

from worker.cli.cmd_backup import cmd_backup
from worker.cli.cmd_cancel import cmd_cancel
from worker.cli.cmd_jobs import cmd_jobs
from worker.cli.cmd_node import cmd_node
from worker.cli.cmd_serve import cmd_serve
from worker.cli.cmd_status import cmd_status
from worker.cli.cmd_submit import cmd_submit
from worker.cli.cmd_tasks import cmd_tasks
from worker.cli.cmd_token import cmd_token


def build_parser() -> argparse.ArgumentParser:
    """One subparser per command, each carrying its handler as ``run``."""
    parser = argparse.ArgumentParser(prog="worker")
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("serve").set_defaults(run=cmd_serve)
    node = sub.add_parser("node")
    node.add_argument("--name", required=True)
    node.set_defaults(run=cmd_node)
    tasks = sub.add_parser("tasks")
    tasks.add_argument("--name", required=True)
    tasks.set_defaults(run=cmd_tasks)
    for name, nodes in (("status", False), ("nodes", True)):
        p = sub.add_parser(name)
        p.add_argument("--json", action="store_true")
        p.set_defaults(run=cmd_status, nodes=nodes)
    jobs = sub.add_parser("jobs")
    jobs.add_argument("--producer", required=True)
    jobs.add_argument("--queue", required=True)
    jobs.add_argument("--json", action="store_true")
    jobs.set_defaults(run=cmd_jobs)
    cancel = sub.add_parser("cancel")
    cancel.add_argument("--producer", required=True)
    cancel.add_argument("job_id")
    cancel.set_defaults(run=cmd_cancel)
    submit = sub.add_parser("submit")
    submit.add_argument("--producer", required=True)
    submit.add_argument("file")
    submit.set_defaults(run=cmd_submit)
    token = sub.add_parser("token").add_subparsers(dest="action", required=True).add_parser("add")
    token.add_argument("--kind", required=True, choices=("producer", "node", "admin"))
    token.add_argument("--name", required=True)
    token.set_defaults(run=cmd_token)
    backup = sub.add_parser("backup")
    backup.add_argument("dest")
    backup.set_defaults(run=cmd_backup)
    return parser
