DELETE FROM p.inputs WHERE job_id = :id;
DELETE FROM p.outputs WHERE result_id IN (SELECT result_id FROM results WHERE job_id = :id);
UPDATE jobs SET payloads_deleted = 1 WHERE id = :id
