"""Aggregate queue, node and failure figures for the CLI (spec 13)."""

import json
import sqlite3
from typing import Any


def read_status(conn: sqlite3.Connection, now: float) -> dict[str, Any]:
    """Counts per queue and state, useful and wasted seconds in the last hour, node reports."""
    queues: dict[str, Any] = {}
    for r in conn.execute(
        "SELECT queue, state, count(*) n, min(created) oldest FROM jobs GROUP BY queue, state"
    ):
        q = queues.setdefault(
            r["queue"], {"states": {}, "oldest_queued_s": None, "done_1h": 0, "wasted_1h_s": 0.0}
        )
        q["states"][r["state"]] = r["n"]
        if r["state"] == "queued":
            q["oldest_queued_s"] = now - r["oldest"]
    for r in conn.execute(
        "SELECT j.queue, a.outcome, count(*) n, sum(coalesce(a.wall_s,0)) s FROM attempts a JOIN jobs j ON j.id=a.job_id"
        " WHERE a.ended>? GROUP BY j.queue, a.outcome",
        (now - 3600,),
    ):
        q = queues.setdefault(
            r["queue"], {"states": {}, "oldest_queued_s": None, "done_1h": 0, "wasted_1h_s": 0.0}
        )
        if r["outcome"] == "succeeded":
            q["done_1h"] = r["n"]
        elif r["outcome"] == "preempted":
            q["wasted_1h_s"] = r["s"]
    nodes = {
        r["name"]: {**json.loads(r["report"]), "age_s": now - r["updated"]}
        for r in conn.execute("SELECT * FROM nodes")
    }
    failures = [
        dict(r)
        for r in conn.execute(
            "SELECT id, queue, error, finished FROM jobs WHERE state='failed' ORDER BY finished DESC LIMIT 10"
        )
    ]
    return {"queues": queues, "nodes": nodes, "recent_failures": failures}
