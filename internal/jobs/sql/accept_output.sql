UPDATE jobs SET state = 'succeeded', error = NULL, finished = :now, updated = :now,
  lease_attempt = NULL, lease_expires = NULL
WHERE id = :id
