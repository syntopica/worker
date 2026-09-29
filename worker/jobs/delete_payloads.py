"""Remove a job's content from the payload file (secure_delete overwrites it)."""

import sqlite3


def delete_payloads(conn: sqlite3.Connection, job_id: str) -> None:
    """Inputs and every output of the job; metadata rows stay, marked as emptied."""
    conn.execute("DELETE FROM p.inputs WHERE job_id=?", (job_id,))
    conn.execute(
        "DELETE FROM p.outputs WHERE result_id IN (SELECT result_id FROM results WHERE job_id=?)",
        (job_id,),
    )
    conn.execute("UPDATE jobs SET payloads_deleted=1 WHERE id=?", (job_id,))
