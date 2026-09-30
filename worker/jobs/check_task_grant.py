"""Whether a queue's grant covers a task's profile, runner, privacy and inputs (spec 8)."""

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.submit_request import SubmitRequest


def check_task_grant(config: WorkerConfig, req: SubmitRequest) -> str:
    """Return the profile name; raise ApiError for every refusal.

    A producer never names a runner the profile does not use, a privacy class
    the profile does not accept, or inputs where the profile has no root.
    """
    name = str(req.input["profile"])
    profile = config.profiles.get(name)
    if profile is None or name not in config.queues[req.queue].profiles:
        raise ApiError(403, "profile_not_granted")
    if req.input["runner"] != profile.runner:
        raise ApiError(400, "runner_mismatch")
    if req.privacy not in profile.privacy:
        raise ApiError(403, "privacy_not_allowed")
    if req.input.get("inputs") and profile.input_root is None:
        raise ApiError(400, "inputs_not_allowed")
    return name
