"""`worker token add`."""

import argparse
import sys

from worker.auth.add_principal import add_principal
from worker.cli.resolve_paths import resolve_paths


def cmd_token(args: argparse.Namespace) -> int:
    """Print the new token once; it is also saved to state/tokens/<name>.token (0600)."""
    _, state = resolve_paths()
    try:
        token = add_principal(state, args.kind, args.name)
    except ValueError as error:
        print(f"worker: {error}", file=sys.stderr)
        return 2
    print(token)
    return 0
