"""Parse ``config.json`` into a WorkerConfig, applying the spec's defaults."""

import json
from pathlib import Path

from worker.config.default_min_free_pct import DEFAULT_MIN_FREE_PCT
from worker.config.default_privacy import DEFAULT_PRIVACY
from worker.config.default_trust import DEFAULT_TRUST
from worker.config.is_loopback_host import is_loopback_host
from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.config.parse_queue_policy import parse_queue_policy
from worker.config.parse_task_profile import parse_task_profile
from worker.config.worker_config import WorkerConfig


def load_worker_config(path: Path) -> WorkerConfig:
    """Return the parsed configuration; raise ValueError on an invalid value."""
    raw = json.loads(path.read_text())
    host, port = str(raw.get("listen", "127.0.0.1:8765")).rsplit(":", 1)
    if not is_loopback_host(host):
        # Tokens travel in plaintext; only a loopback bind keeps them on the host.
        raise ValueError(f"listen: {host} is not a loopback address")
    privacy = dict(DEFAULT_PRIVACY)
    trust = dict(DEFAULT_TRUST)
    for name, override in (raw.get("privacy") or {}).items():
        privacy[name] = frozenset(override["executors"])
        trust[name] = frozenset(override["trust"])
    return WorkerConfig(
        listen_host=host,
        listen_port=int(port),
        models={
            n: ModelPin(
                n, int(m["num_ctx"]), str(m["keep_alive"]), float(m["cold_gb"]), float(m["warm_gb"])
            )
            for n, m in raw["models"].items()
        },
        queues={n: parse_queue_policy(n, q) for n, q in raw["queues"].items()},
        nodes={
            n: NodePolicy(
                n,
                v["trust"],
                float(v.get("idle_threshold_s", 300)),
                float(v["memory_budget_gb"]),
                v["ollama_url"],
                v["ollama_launchd_label"],
                float(v.get("min_free_pct", DEFAULT_MIN_FREE_PCT)),
            )
            for n, v in raw["nodes"].items()
        },
        producers={n: frozenset(q) for n, q in raw["producers"].items()},
        privacy=privacy,
        trust=trust,
        max_payload_bytes=int(raw.get("max_payload_bytes", 1_048_576)),
        max_split=int(raw.get("max_split", 8)),
        max_parked_runs=int(raw.get("max_parked_runs", 3)),
        split_after_preemptions=int(raw.get("split_after_preemptions", 3)),
        profiles={
            n: parse_task_profile(n, p, path.parent) for n, p in (raw.get("profiles") or {}).items()
        },
        runner_cooldown_s={
            n: float(r.get("cooldown_s", 3600)) for n, r in (raw.get("runners") or {}).items()
        },
    )
