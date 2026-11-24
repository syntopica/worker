UPDATE results SET acked = :now
WHERE job_id = :id AND control = 'split_requested' AND acked IS NULL
