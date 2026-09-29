"""Everything a route handler needs for one request."""

import sqlite3
from dataclasses import dataclass

from worker.auth.principal import Principal
from worker.config.worker_config import WorkerConfig


@dataclass(frozen=True)
class RequestContext:
    """``parts`` is the path split on ``/`` without empty segments."""

    principal: Principal
    method: str
    parts: list[str]
    query: dict[str, str]
    body: bytes
    config: WorkerConfig
    conn: sqlite3.Connection
    now: float
