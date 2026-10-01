"""Build a queue's runner route from its raw configuration entry."""

from typing import Any

from worker.config.runner_route import RunnerRoute


def parse_runner_route(name: str, raw: dict[str, Any] | None) -> RunnerRoute | None:
    """None when the queue has no route; raise ValueError on a malformed one."""
    if raw is None:
        return None
    profile = raw.get("profile")
    if not isinstance(profile, str) or not profile:
        raise ValueError(f"queue {name}: runner.profile must name a task profile")
    return RunnerRoute(profile, float(raw.get("after_s", 0)))
