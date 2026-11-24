UPDATE jobs SET acked = :now, updated = :now WHERE id = :id AND acked IS NULL
