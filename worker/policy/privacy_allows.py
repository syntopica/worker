"""Whether a privacy class may use an executor on a node of a trust class (spec 8)."""

from worker.config.worker_config import WorkerConfig


def privacy_allows(config: WorkerConfig, privacy: str, executor: str, trust: str) -> bool:
    """Both the executor and the node trust must be permitted; absence denies."""
    return executor in config.privacy.get(privacy, frozenset()) and trust in config.trust.get(
        privacy, frozenset()
    )
