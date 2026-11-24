UPDATE jobs SET state = 'queued', error = :code, not_before = :due, updated = :now,
  lease_attempt = NULL, lease_expires = NULL
WHERE id = :id
