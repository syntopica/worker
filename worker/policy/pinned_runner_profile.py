"""The runner profile a pinned shadow member or judge may run under now."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.candidate import Candidate
from worker.policy.privacy_allows import privacy_allows
from worker.policy.profile_resting import profile_resting


def pinned_runner_profile(
    c: Candidate, config: WorkerConfig, cooling: frozenset[str], trust: str
) -> str | None:
    """A ``runner:`` pin's profile, or a judge's own, when usable now; else None.

    Usable means configured, its runner rested, and the job's class allowed by
    the profile and for ``runner`` executors, checked at lease time so a
    configuration change after the group opened still applies.
    """
    if c.pin == "judge":
        name = c.model
    else:
        kind, _, name = (c.pin or "").partition(":")
        if kind != "runner":
            return None
    profile = config.profiles.get(name)
    if profile is None or profile.on_demand or profile_resting(profile, cooling):
        return None
    if c.privacy not in profile.privacy:
        return None
    return name if privacy_allows(config, c.privacy, "runner", trust) else None
