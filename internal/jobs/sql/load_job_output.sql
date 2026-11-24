SELECT o.body FROM results r JOIN p.outputs o ON o.result_id = r.result_id
WHERE r.job_id = :id ORDER BY r.seq DESC LIMIT 1
