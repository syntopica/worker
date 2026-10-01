"""Render an inference job's input as the task input a runner understands."""

from typing import Any


def inference_task_input(job_input: dict[str, Any]) -> dict[str, Any]:
    """The messages become one prompt, each under its role; ``schema`` the output schema."""
    prompt = "\n\n".join(f"[{m['role']}]\n{m['content']}" for m in job_input["messages"])
    task: dict[str, Any] = {"prompt": prompt}
    if job_input.get("schema") is not None:
        task["output_schema"] = job_input["schema"]
    return task
