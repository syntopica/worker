"""Whether a privacy class may be sent to a shadow target or a runner profile."""

from worker.config.worker_config import WorkerConfig


def target_allowed(config: WorkerConfig, privacy: str, target: str) -> bool:
    """The class must permit the target's executor; a runner profile must also list it."""
    kind, _, name = target.partition(":")
    if kind not in config.privacy.get(privacy, frozenset()):
        return False
    if kind != "runner":
        return True
    profile = config.profiles.get(name)
    return profile is not None and privacy in profile.privacy
