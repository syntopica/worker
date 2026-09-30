"""`worker serve` - the coordinator."""

import argparse
import errno
import sys
import threading
import time

from worker.api.build_server import build_server
from worker.cli.resolve_paths import resolve_paths
from worker.cli.run_sweeper import run_sweeper
from worker.config.config_source import ConfigSource
from worker.config.load_worker_config import load_worker_config
from worker.jobs.reconcile_lost_payloads import reconcile_lost_payloads
from worker.store.migrate_state import migrate_state
from worker.store.open_store import open_store


def cmd_serve(args: argparse.Namespace) -> int:  # noqa: ARG001
    """Migrate the store, fail jobs whose payload is gone, then serve; reclaim and retention run every minute."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    source = ConfigSource(config_path, config)
    migrate_state(state)
    conn = open_store(state)
    try:
        reconcile_lost_payloads(conn, time.time())
    finally:
        conn.close()
    try:
        server = build_server(config, state, source.current)
    except OSError as error:
        if error.errno != errno.EADDRINUSE:
            raise
        print(
            f"worker: {config.listen_host}:{config.listen_port} is already in use", file=sys.stderr
        )
        return 2
    threading.Thread(
        target=run_sweeper, args=(state, config), kwargs={"current": source.current}, daemon=True
    ).start()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()
    return 0
