UPDATE jobs SET state = :state, lease_expires = :expires, updated = :now WHERE id = :id
