SELECT result_id, control FROM results WHERE job_id = :id AND acked IS NULL ORDER BY seq
