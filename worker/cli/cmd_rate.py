"""`worker rate`: rate a result a script already collected."""

import argparse
import json

from worker.cli.build_client import build_client


def cmd_rate(args: argparse.Namespace) -> int:
    """Ack the result again with ``good``, ``edited`` or ``discarded``.

    ``worker run`` acks as it collects and names the result on stderr; the
    script judges the output afterwards, and a later ack may change the rating.
    """
    build_client(args.producer).ack(args.job_id, args.result_id, rating=args.rating)
    print(json.dumps({"rated": args.rating}))
    return 0
