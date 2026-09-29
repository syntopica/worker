"""Admit a split child: validate it against its parent and supersede the parent."""

import re
import sqlite3

from worker.config.worker_config import WorkerConfig
from worker.jobs.api_error import ApiError
from worker.jobs.submit_request import SubmitRequest

_CHILD_KEY = re.compile(r"/(0|[1-9]\d*)/([1-9]\d*)$")
_MIN_SPLIT = 2


def admit_split_child(
    conn: sqlite3.Connection, config: WorkerConfig, producer: str, req: SubmitRequest, now: float
) -> None:
    """The key is exactly ``<parent key>/<index>/<count>``; then mark the parent superseded."""
    parent = conn.execute(
        "SELECT producer, queue, state, split_count, idempotency_key FROM jobs WHERE id=?",
        (req.parent_id,),
    ).fetchone()
    match = _CHILD_KEY.search(req.idempotency_key)
    if (
        parent is None
        or parent["producer"] != producer
        or parent["queue"] != req.queue
        or match is None
        or req.idempotency_key
        != f"{parent['idempotency_key']}/{int(match.group(1))}/{int(match.group(2))}"
    ):
        raise ApiError(400, "bad_split_child")
    if parent["state"] not in ("split_requested", "superseded"):
        raise ApiError(409, "parent_not_splitting")
    count = int(match.group(2))
    if not _MIN_SPLIT <= count <= config.max_split or int(match.group(1)) >= count:
        raise ApiError(400, "bad_split_count")
    if parent["split_count"] is not None and parent["split_count"] != count:
        raise ApiError(409, "split_count_mismatch")
    conn.execute(
        "UPDATE jobs SET state='superseded', split_count=?, finished=coalesce(finished, ?), updated=? WHERE id=?",
        (count, now, now, req.parent_id),
    )
