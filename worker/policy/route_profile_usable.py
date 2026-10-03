"""Whether one task profile can take a routed inference job right now."""

from worker.config.task_profile import TaskProfile
from worker.policy.profile_resting import profile_resting


def route_profile_usable(
    profile: TaskProfile | None, privacy: str, cooling: frozenset[str], node: str
) -> bool:
    """Configured, not resting, allowed on ``node`` and for the job's class."""
    if profile is None or profile_resting(profile, cooling):
        return False
    if profile.nodes is not None and node not in profile.nodes:
        return False
    return privacy in profile.privacy
