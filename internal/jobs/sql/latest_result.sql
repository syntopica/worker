SELECT r.seq, r.result_id, r.job_id, r.control, r.detail, o.body, r.executor, r.usage
FROM results r LEFT JOIN p.outputs o ON o.result_id = r.result_id
WHERE r.producer = :producer AND r.job_id = :id AND r.acked IS NULL
ORDER BY r.seq DESC LIMIT 1
