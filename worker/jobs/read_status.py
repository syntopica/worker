"""Aggregate queue, node and failure figures for the CLI (spec 13)."""

import json
import sqlite3
from typing import Any


def read_status(conn: sqlite3.Connection, now: float) -> dict[str, Any]:
    """Counts per queue and state, useful and wasted seconds in the last hour, node reports.

    ``sampling_failed`` is how many of a queue's failed jobs were shadow copies or judges.

    Each node also carries its last release in a day (preemption code and age), and
    ``cooldowns`` maps each resting runner to the seconds left. Recent failures are
    production jobs only: shadow copies and their judges are sampling, not lost work.
    """
    queues: dict[str, Any] = {}
    for r in conn.execute(
        "SELECT queue, state, count(*) n, min(created) oldest,"
        " sum(producer IN ('_shadow', '_judge')) sampling FROM jobs GROUP BY queue, state"
    ):
        q = queues.setdefault(
            r["queue"],
            {
                "states": {},
                "oldest_queued_s": None,
                "done_1h": 0,
                "wasted_1h_s": 0.0,
                "sampling_failed": 0,
            },
        )
        q["states"][r["state"]] = r["n"]
        if r["state"] == "queued":
            q["oldest_queued_s"] = now - r["oldest"]
        if r["state"] == "failed":
            q["sampling_failed"] = r["sampling"]
    for r in conn.execute(
        "SELECT j.queue, a.outcome, count(*) n, sum(coalesce(a.wall_s,0)) s FROM attempts a JOIN jobs j ON j.id=a.job_id"
        " WHERE a.ended>? GROUP BY j.queue, a.outcome",
        (now - 3600,),
    ):
        q = queues.setdefault(
            r["queue"],
            {
                "states": {},
                "oldest_queued_s": None,
                "done_1h": 0,
                "wasted_1h_s": 0.0,
                "sampling_failed": 0,
            },
        )
        if r["outcome"] == "succeeded":
            q["done_1h"] = r["n"]
        elif r["outcome"] == "preempted":
            q["wasted_1h_s"] = r["s"]
    nodes = {
        r["name"]: {**json.loads(r["report"]), "age_s": now - r["updated"]}
        for r in conn.execute("SELECT * FROM nodes")
    }
    for r in conn.execute(
        "SELECT node, error, max(ended) ended FROM attempts"
        " WHERE started>? AND outcome='preempted' GROUP BY node",
        (now - 86400,),
    ):
        if r["node"] in nodes:
            nodes[r["node"]]["last_release"] = {"code": r["error"], "age_s": now - r["ended"]}
    cooldowns = {
        r["runner"]: r["until"] - now
        for r in conn.execute("SELECT runner, until FROM cooldowns WHERE until>?", (now,))
    }
    failures = [
        dict(r)
        for r in conn.execute(
            "SELECT id, queue, error, finished FROM jobs WHERE state='failed'"
            " AND producer NOT IN ('_shadow', '_judge') ORDER BY finished DESC LIMIT 10"
        )
    ]
    return {
        "queues": queues,
        "nodes": nodes,
        "cooldowns": cooldowns,
        "recent_failures": failures,
    }
