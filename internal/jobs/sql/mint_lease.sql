UPDATE jobs SET state = 'leased', generation = generation + 1,
  lease_node = :node, lease_attempt = :attempt, lease_expires = :expires, updated = :now,
  model = CASE WHEN kind = 'task' AND :profile IS NULL THEN :model ELSE model END
WHERE id = :id
RETURNING generation
