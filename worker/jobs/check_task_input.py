"""Validate the shape of a task ``input`` before it is stored (amendment 2026-09-30)."""

from typing import Any

from worker.jobs.api_error import ApiError
from worker.jobs.is_safe_relative_path import is_safe_relative_path

_MAX_INPUTS = 64
_KEYS = {"runner", "profile", "prompt", "inputs", "output_schema"}


def check_task_input(job_input: dict[str, Any]) -> None:
    """Raise ApiError(400, "bad_input") unless it is a runnable read-only task."""
    if set(job_input) - _KEYS:
        raise ApiError(400, "bad_input")
    for name in ("profile", "prompt"):
        if not isinstance(job_input.get(name), str) or not job_input[name]:
            raise ApiError(400, "bad_input")
    if "runner" in job_input and not isinstance(job_input["runner"], str):
        raise ApiError(400, "bad_input")
    inputs = job_input.get("inputs", [])
    if not isinstance(inputs, list) or len(inputs) > _MAX_INPUTS:
        raise ApiError(400, "bad_input")
    if not all(is_safe_relative_path(entry) for entry in inputs):
        raise ApiError(400, "bad_input")
    if "output_schema" in job_input and not isinstance(job_input["output_schema"], dict):
        raise ApiError(400, "bad_input")
