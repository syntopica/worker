UPDATE jobs SET payloads_deleted = 1
WHERE finished IS NOT NULL AND id NOT IN (SELECT job_id FROM p.inputs)
