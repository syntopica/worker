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
    fallbacks = raw.get("fallbacks", [])
    if not isinstance(fallbacks, list) or not all(isinstance(f, str) and f for f in fallbacks):
        raise ValueError(f"queue {name}: runner.fallbacks must list task profiles")
    return RunnerRoute(profile, float(raw.get("after_s", 0)), tuple(fallbacks))
