SELECT id FROM jobs WHERE state = 'succeeded' AND acked IS NULL AND queue = :queue
  AND ((privacy IN ('mail', 'personal', 'secret') AND finished <= :unacked_before)
       OR (privacy NOT IN ('mail', 'personal', 'secret') AND finished <= :retained_before))
LIMIT :limit
