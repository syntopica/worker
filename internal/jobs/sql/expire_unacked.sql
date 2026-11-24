UPDATE jobs SET state = 'unacked_expired', updated = :now WHERE id = :id;
UPDATE results SET acked = :now WHERE job_id = :id AND acked IS NULL
