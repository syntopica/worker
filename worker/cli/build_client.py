"""A producer client for the configured coordinator."""

from worker.cli.base_url import base_url
from worker.cli.read_token import read_token
from worker.cli.resolve_paths import resolve_paths
from worker.client.worker_client import WorkerClient
from worker.config.load_worker_config import load_worker_config


def build_client(principal: str) -> WorkerClient:
    """Client authenticated with state/tokens/<principal>.token."""
    config_path, state = resolve_paths()
    config = load_worker_config(config_path)
    return WorkerClient(base_url(config), read_token(state, principal))
