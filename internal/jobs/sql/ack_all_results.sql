UPDATE results SET acked = :now WHERE job_id = :id AND acked IS NULL
