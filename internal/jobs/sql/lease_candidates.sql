SELECT id, queue, model, priority, privacy, created, parked, parked_min_idle_s, kind,
  coalesce(pin, ''), coalesce(error, ''), coalesce(exclude_providers, '')
FROM (
  SELECT *, row_number() OVER (PARTITION BY queue, kind ORDER BY priority DESC, created) AS rank
  FROM jobs
  WHERE state = 'queued' AND kind IN ('inference', :extra_kind) AND not_before <= :now
    AND (:created_before IS NULL OR created <= :created_before)
    AND (deadline IS NULL OR deadline > :now)
    AND privacy IN (SELECT value FROM json_each(:classes))
    AND queue IN (SELECT value FROM json_each(:queues))
)
WHERE rank <= 200
ORDER BY priority DESC, created
