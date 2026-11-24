UPDATE jobs SET state = 'queued', lease_attempt = NULL, lease_expires = NULL,
  not_before = :now, updated = :now WHERE id = :id
