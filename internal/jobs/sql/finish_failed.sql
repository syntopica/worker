UPDATE jobs SET state = :state, error = :code, finished = :now, updated = :now, lease_attempt = NULL,
  lease_expires = NULL
WHERE id = :id
