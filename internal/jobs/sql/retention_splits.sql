SELECT id FROM jobs
WHERE state = 'split_requested' AND queue = :queue AND updated <= :unacked_before
LIMIT :limit
