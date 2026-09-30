"""Move a job to a terminal failure and tell its producer."""

import sqlite3
from typing import Any

from worker.jobs.add_control_result import add_control_result


def finish_failed(  # noqa: PLR0913
    conn: sqlite3.Connection,
    job: sqlite3.Row,
    code: str,
    now: float,
    control: str = "failed",
    *,
    extra: dict[str, Any] | None = None,
) -> None:
    """``control`` is ``failed`` or ``expired``; ``code`` is an allowlisted error code.

    ``extra`` adds allowlisted detail fields (``schema_path``) to the control result.
    """
    conn.execute(
        "UPDATE jobs SET state=?, error=?, finished=?, updated=?, lease_attempt=NULL, lease_expires=NULL WHERE id=?",
        (control, code, now, now, job["id"]),
    )
    add_control_result(conn, job, control, {"error": code, **(extra or {})}, now)
