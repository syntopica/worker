"""The runners whose quota CodexBar reports spent, and until when."""

from collections.abc import Callable
from typing import Any

from worker.config.worker_config import WorkerConfig
from worker.tasks.model_window_walls import model_window_walls
from worker.tasks.read_codexbar_usage import read_codexbar_usage
from worker.tasks.spent_until import spent_until


def probe_runner_walls(
    config: WorkerConfig,
    now: float,
    read: Callable[[str], dict[str, Any] | None] = read_codexbar_usage,
) -> dict[str, float]:
    """``{runner: until}`` for each configured runner known to be spent now.

    A runner with ``model_windows`` rests per model instead (``runner:model``).
    """
    walls: dict[str, float] = {}
    for runner, (provider, windows) in config.runner_quota.items():
        usage = read(provider)
        if usage is None:
            continue
        model_windows = config.runner_model_windows.get(runner)
        if model_windows:
            walls.update(model_window_walls(runner, usage, model_windows, config.profiles, now))
            continue
        until = spent_until(usage, windows, now)
        if until is not None:
            walls[runner] = until
    return walls
