"""Indexes for the hot paths, created on every open so an older store gains them."""

INDEXES = (
    "CREATE INDEX IF NOT EXISTS results_job ON results (job_id)",
    "CREATE INDEX IF NOT EXISTS results_feed ON results (producer, queue, acked, seq)",
    "CREATE INDEX IF NOT EXISTS attempts_job ON attempts (job_id)",
    "CREATE INDEX IF NOT EXISTS attempts_ended ON attempts (ended)",
    "CREATE INDEX IF NOT EXISTS jobs_queue ON jobs (queue)",
    "CREATE INDEX IF NOT EXISTS jobs_outstanding ON jobs (producer, queue, state)",
    "CREATE INDEX IF NOT EXISTS jobs_candidates ON jobs (state, priority DESC, created)",
    "CREATE INDEX IF NOT EXISTS jobs_unacked ON jobs (state, acked, queue, finished)",
    "CREATE INDEX IF NOT EXISTS jobs_payloads ON jobs (payloads_deleted, queue, finished)",
    "CREATE INDEX IF NOT EXISTS jobs_release ON jobs (payloads_deleted, queue, updated)",
    "CREATE INDEX IF NOT EXISTS jobs_shadow ON jobs (shadow_of)",
    "CREATE INDEX IF NOT EXISTS judgements_feed ON judgements (queue, created)",
)
