"""The coordinator's HTTP server, bound to the configured address only."""

from http.server import ThreadingHTTPServer
from pathlib import Path

from worker.api.make_handler import make_handler
from worker.config.worker_config import WorkerConfig


def build_server(config: WorkerConfig, state_dir: Path) -> ThreadingHTTPServer:
    """Loopback in phase 1 (spec 11); ``listen`` port 0 picks a free port for tests."""
    return ThreadingHTTPServer(
        (config.listen_host, config.listen_port), make_handler(config, state_dir)
    )
