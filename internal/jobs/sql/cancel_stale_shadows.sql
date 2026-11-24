UPDATE jobs SET state = 'cancelled', finished = :now, updated = :now
WHERE producer IN ('_shadow', '_judge') AND state IN ('queued', 'split_requested') AND created < :since
