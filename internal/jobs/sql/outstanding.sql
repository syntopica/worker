SELECT count(*) FROM jobs
WHERE producer = :producer AND queue = :queue AND acked IS NULL
  AND state NOT IN ('superseded', 'cancelled', 'unacked_expired')
