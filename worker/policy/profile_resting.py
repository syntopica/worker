"""Whether a task profile is resting after a quota wall."""

from worker.config.cooldown_key import cooldown_key
from worker.config.task_profile import TaskProfile


def profile_resting(profile: TaskProfile, cooling: frozenset[str]) -> bool:
    """Its whole runner rests, or the model it pins does."""
    return profile.runner in cooling or cooldown_key(profile.runner, profile.model) in cooling
