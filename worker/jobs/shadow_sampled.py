"""Whether a job falls in its queue's shadow sample."""

import hashlib


def shadow_sampled(job_id: str, rate: float) -> bool:
    """Deterministic in the job id, so a retried completion decides the same way."""
    bucket = int(hashlib.sha256(job_id.encode()).hexdigest()[:8], 16)
    return bucket < rate * 0x100000000
