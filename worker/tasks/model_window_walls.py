"""Per-model rests for a runner whose provider meters model families apart."""

from collections.abc import Mapping
from typing import Any

from worker.config.cooldown_key import cooldown_key
from worker.config.task_profile import TaskProfile
from worker.tasks.spent_until import spent_until


def model_window_walls(
    runner: str,
    usage: dict[str, Any],
    windows: Mapping[str, tuple[str, ...]],
    profiles: Mapping[str, TaskProfile],
    now: float,
) -> dict[str, float]:
    """``{runner:model: until}`` for each profile model a spent window meters.

    ``windows`` maps a window label to the model-name fragments it meters, so
    a spent Gemini window rests the runner's Gemini models and leaves the
    others it serves from a separate allowance free to run.
    """
    walls: dict[str, float] = {}
    for window, fragments in windows.items():
        until = spent_until(usage, (window,), now)
        if until is None:
            continue
        for profile in profiles.values():
            model = (profile.model or "").lower()
            if profile.runner == runner and model and any(f.lower() in model for f in fragments):
                key = cooldown_key(runner, profile.model)
                walls[key] = max(walls.get(key, until), until)
    return walls
