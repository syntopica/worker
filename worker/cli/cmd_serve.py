"""`worker serve` - the coordinator."""

import argparse
import threading

from worker.api.build_server import build_server
from worker.cli.resolve_paths import resolve_paths
from worker.cli.run_sweeper import run_sweeper
from worker.config.load_worker_config import load_worker_config


def cmd_serve(args: argparse.Namespace) -> int:  # noqa: ARG001
    """Serve until interrupted; lease reclaim and retention run every minute."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    server = build_server(config, state)
    threading.Thread(target=run_sweeper, args=(state, config), daemon=True).start()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
    return 0
