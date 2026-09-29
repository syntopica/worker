"""Make a job id safe to put in a URL path."""

import urllib.parse


def quote_job_id(job_id: str) -> str:
    """Percent-encode everything, including ``/``."""
    return urllib.parse.quote(job_id, safe="")
