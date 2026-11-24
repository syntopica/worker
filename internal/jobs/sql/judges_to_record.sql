SELECT j.id, j.shadow_of, j.queue, o.body
FROM jobs j JOIN results r ON r.job_id = j.id JOIN p.outputs o ON o.result_id = r.result_id
JOIN p.inputs i ON i.job_id = j.id
WHERE j.pin = 'judge' AND j.state = 'succeeded'
  AND j.shadow_of NOT IN (SELECT group_id FROM judgements)
  AND EXISTS (SELECT 1 FROM json_each(i.body, '$.input.labels'))
LIMIT :limit
