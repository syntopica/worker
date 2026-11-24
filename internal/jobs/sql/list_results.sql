SELECT r.seq, r.result_id, r.job_id, r.control, r.detail, o.body, r.executor, r.usage
FROM results r LEFT JOIN p.outputs o ON o.result_id = r.result_id
WHERE r.producer = :producer AND r.queue = :queue AND r.seq > :after AND r.acked IS NULL
  AND (r.control IS NOT NULL OR o.result_id IS NOT NULL)
ORDER BY r.seq LIMIT :limit
