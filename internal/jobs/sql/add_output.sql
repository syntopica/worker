INSERT INTO results (result_id, job_id, producer, queue, executor, usage, created)
VALUES (:result_id, :id, :producer, :queue, :executor, :usage, :now);
INSERT INTO p.outputs (result_id, body) VALUES (:result_id, :body)
