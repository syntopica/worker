"""The SELECT every admin job read starts from: metadata columns and the attempt count."""

ADMIN_JOB_SELECT = (
    "SELECT j.id, j.queue, j.producer, j.state, j.privacy, j.tier, j.created, j.updated,"
    " j.error, j.acked, j.retry_of,"
    " (SELECT count(*) FROM attempts a WHERE a.job_id=j.id) attempt_count FROM jobs j"
)
