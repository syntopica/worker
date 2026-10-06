"""Build one TaskProfile from its raw configuration entry."""

from pathlib import Path
from typing import Any

from worker.config.task_profile import TaskProfile
from worker.config.task_runners import TASK_RUNNERS


def parse_task_profile(name: str, raw: dict[str, Any], config_dir: Path) -> TaskProfile:
    """Apply the defaults; raise ValueError on an unknown runner.

    ``input_root`` resolves against the directory of ``config.json``, the rule
    every Syntopica path follows. An agy profile takes no inputs: reading files
    needs ``--dangerously-skip-permissions``, a later and explicit decision.
    A max-lane profile has no tools either, pins its model, and is on demand:
    it spends a subscription the owner releases explicitly.
    """
    runner = raw.get("runner")
    if runner not in TASK_RUNNERS:
        raise ValueError(f"profile {name}: runner must be one of {TASK_RUNNERS}")
    root = raw.get("input_root")
    if root is not None and runner == "agy":
        raise ValueError(f"profile {name}: agy profiles take no input_root")
    if runner == "max-lane" and (root is not None or not raw.get("model")):
        raise ValueError(f"profile {name}: max-lane profiles name a model and take no input_root")
    if runner == "max-lane" and raw.get("on_demand") is not True:
        raise ValueError(f"profile {name}: max-lane profiles are on_demand")
    nodes = raw.get("nodes")
    return TaskProfile(
        name=name,
        runner=runner,
        model=raw.get("model"),
        reasoning=raw.get("reasoning"),
        privacy=frozenset(raw.get("privacy", ("public", "internal"))),
        timeout_s=float(raw.get("timeout_s", 900)),
        nodes=None if nodes is None else frozenset(nodes),
        input_root=None if root is None else (config_dir / str(root)).resolve(),
        command=raw.get("command"),
        env_unset=frozenset(raw.get("env_unset", ())),
        denied_root=(config_dir / "state").resolve(),
        on_demand=raw.get("on_demand") is True,
    )
