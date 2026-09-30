"""The node's own privacy check of a lease before it runs anything (spec 8)."""

from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.policy.privacy_allows import privacy_allows


def lease_privacy_allowed(
    config: WorkerConfig, node_name: str, lease: dict[str, Any], executor: str
) -> bool:
    """Whether this node's configuration lets the lease's class use ``executor`` here.

    The coordinator already filtered by the same policy; this catches a
    coordinator and node that disagree (a stale or edited config). A lease
    without a privacy class is refused: absence denies.
    """
    trust = config.nodes[node_name].trust
    return privacy_allows(config, str(lease.get("privacy", "")), executor, trust)
