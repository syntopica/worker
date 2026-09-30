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
    """
    runner = raw.get("runner")
    if runner not in TASK_RUNNERS:
        raise ValueError(f"profile {name}: runner must be one of {TASK_RUNNERS}")
    root = raw.get("input_root")
    if root is not None and runner == "agy":
        raise ValueError(f"profile {name}: agy profiles take no input_root")
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
    )
