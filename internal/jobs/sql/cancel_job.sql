UPDATE jobs SET state = 'cancelled', acked = :now, finished = :now, updated = :now,
  lease_attempt = NULL, lease_expires = NULL
WHERE id = :id
