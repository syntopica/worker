"""Whether a shadow target names the executor that already answered."""

from typing import Any

from worker.config.worker_config import WorkerConfig


def target_answered(config: WorkerConfig, target: str, executor: dict[str, Any]) -> bool:
    """``ollama:``/``openrouter:`` match provider and model; ``runner:`` matches the runner."""
    kind, _, name = target.partition(":")
    provider = str(executor.get("provider") or "")
    if kind == "runner":
        profile = config.profiles.get(name)
        return profile is not None and profile.runner == provider
    return kind == provider and name == str(executor.get("model") or "")
