"""The runners already running as many attempts as their ``max_concurrent`` allows."""

import sqlite3

from worker.config.worker_config import WorkerConfig

_RUNNER_PIN = "runner:"


def saturated_runners(conn: sqlite3.Connection, config: WorkerConfig) -> frozenset[str]:
    """Count live tasks by their profile and pinned inference by its ``runner:`` pin.

    The coordinator counts across every node, so a cap holds however many
    task slots the nodes run.
    """
    if not config.runner_max_concurrent:
        return frozenset()
    live: dict[str, int] = {}
    for kind, model, pin in conn.execute(
        "SELECT kind, model, pin FROM jobs WHERE state IN ('leased','running','draining')"
    ):
        if kind == "task":
            name = model
        elif (pin or "").startswith(_RUNNER_PIN):
            name = pin[len(_RUNNER_PIN) :]
        else:
            continue
        profile = config.profiles.get(name)
        if profile is not None:
            live[profile.runner] = live.get(profile.runner, 0) + 1
    return frozenset(
        runner for runner, cap in config.runner_max_concurrent.items() if live.get(runner, 0) >= cap
    )
