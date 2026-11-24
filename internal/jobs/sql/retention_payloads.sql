SELECT id FROM jobs WHERE payloads_deleted = 0 AND queue = :queue
  AND finished <= max(:inspected_before, :retained_before)
  AND NOT (state = 'succeeded' AND acked IS NULL)
  AND ((privacy IN ('mail', 'personal', 'secret') AND state != 'succeeded' AND finished <= :inspected_before)
       OR coalesce(acked, finished) <= :retained_before)
LIMIT :limit
