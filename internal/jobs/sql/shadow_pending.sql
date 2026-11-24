SELECT count(*) FROM jobs WHERE queue = :queue AND producer = '_shadow'
  AND state IN ('queued', 'leased', 'running', 'draining', 'split_requested')
