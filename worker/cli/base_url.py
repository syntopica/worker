"""The coordinator address from the configuration."""

from worker.config.worker_config import WorkerConfig

_WILDCARDS = {"0.0.0.0": "127.0.0.1", "::": "[::1]"}  # noqa: S104 - mapped away, never bound


def base_url(config: WorkerConfig) -> str:
    """``http://host:port``; a wildcard listen host is reached over loopback."""
    host = _WILDCARDS.get(config.listen_host, config.listen_host)
    if ":" in host and not host.startswith("["):
        host = f"[{host}]"
    return f"http://{host}:{config.listen_port}"
