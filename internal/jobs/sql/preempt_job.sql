UPDATE jobs SET state = :state, preemptions = :preemptions, updated = :now,
  lease_attempt = NULL, lease_expires = NULL
WHERE id = :id
