"""The whole instance configuration, parsed and defaulted."""

from collections.abc import Mapping
from dataclasses import dataclass

from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.config.queue_policy import QueuePolicy


@dataclass(frozen=True)
class WorkerConfig:
    """``producers`` maps a producer to the queues it is granted."""

    listen_host: str
    listen_port: int
    models: Mapping[str, ModelPin]
    queues: Mapping[str, QueuePolicy]
    nodes: Mapping[str, NodePolicy]
    producers: Mapping[str, frozenset[str]]
    privacy: Mapping[str, frozenset[str]]
    trust: Mapping[str, frozenset[str]]
    max_payload_bytes: int
    max_split: int
    max_parked_runs: int
    split_after_preemptions: int
