"""Entry point: `worker <command>`."""

import sys
import urllib.error

from worker.cli.build_parser import build_parser
from worker.cli.missing_token_error import MissingTokenError
from worker.client.api_failure import ApiFailure


def main(argv: list[str] | None = None) -> int:
    """Parse and dispatch; refusals and a missing token are one stderr line, never a traceback."""
    args = build_parser().parse_args(argv)
    try:
        return int(args.run(args))
    except ApiFailure as failure:
        print(f"worker: {failure.status} {failure.code}", file=sys.stderr)
        return 1
    except MissingTokenError as missing:
        print(f"worker: {missing}", file=sys.stderr)
        return 2
    except urllib.error.URLError as error:
        print(f"worker: coordinator unreachable ({type(error.reason).__name__})", file=sys.stderr)
        return 1
    except (ConnectionError, TimeoutError) as error:
        print(f"worker: coordinator unreachable ({type(error).__name__})", file=sys.stderr)
        return 1
    except (OSError, ValueError) as error:
        # Only the class and a file name: messages may quote file contents or payloads.
        named = getattr(error, "filename", None)
        detail = f": {named}" if named else ""
        print(f"worker: {type(error).__name__}{detail}", file=sys.stderr)
        return 2
