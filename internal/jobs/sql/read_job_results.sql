SELECT result_id, control, detail, executor, usage, rating, created, acked
FROM results
WHERE job_id=:id
ORDER BY seq
