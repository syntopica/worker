SELECT id, queue, error, finished FROM jobs
WHERE state = 'failed' AND producer NOT IN ('_shadow', '_judge')
ORDER BY finished DESC LIMIT 10
