"""The reload-sensitive options one managed model is always called with."""

from dataclasses import dataclass


@dataclass(frozen=True)
class ModelPin:
    """``cold_gb`` covers weights, KV for num_ctx x parallel, buffers and margin."""

    name: str
    num_ctx: int
    keep_alive: str
    cold_gb: float
    warm_gb: float
