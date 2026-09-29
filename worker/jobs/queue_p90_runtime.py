"""The 90th percentile wall time of a queue's successful attempts."""

import sqlite3


def queue_p90_runtime(conn: sqlite3.Connection, queue: str) -> float | None:
    """Over the last 200 successes; None when there is no history yet."""
    walls = sorted(
        r[0]
        for r in conn.execute(
            "SELECT a.wall_s FROM attempts a JOIN jobs j ON j.id=a.job_id"
            " WHERE j.queue=? AND a.outcome='succeeded' ORDER BY a.started DESC LIMIT 200",
            (queue,),
        )
    )
    if not walls:
        return None
    return float(walls[min(len(walls) - 1, int(0.9 * len(walls)))])
