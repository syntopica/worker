SELECT id FROM jobs WHERE payloads_deleted = 1 AND queue = :queue AND updated <= :retained_before
  AND (state IN ('failed', 'expired', 'cancelled', 'superseded', 'unacked_expired', 'reconciled')
       OR (state = 'succeeded' AND acked IS NOT NULL))
LIMIT :limit
