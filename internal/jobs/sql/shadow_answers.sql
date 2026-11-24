SELECT j.id, r.executor, o.body
FROM jobs j JOIN results r ON r.job_id = j.id JOIN p.outputs o ON o.result_id = r.result_id
WHERE j.shadow_of = :group AND j.state = 'succeeded' ORDER BY j.id
