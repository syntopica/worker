"""The coordinator address from the configuration."""

from worker.config.worker_config import WorkerConfig


def base_url(config: WorkerConfig) -> str:
    """``http://host:port`` of the configured listen address."""
    return f"http://{config.listen_host}:{config.listen_port}"
