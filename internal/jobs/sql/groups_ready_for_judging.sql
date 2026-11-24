SELECT shadow_of FROM jobs WHERE shadow_of IS NOT NULL GROUP BY shadow_of
HAVING sum(pin = 'judge') = 0 AND sum(state IN ('queued', 'leased', 'running', 'draining', 'split_requested')) = 0
LIMIT :limit
