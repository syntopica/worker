SELECT coalesce(a.profile,
  CASE WHEN j.kind = 'task' THEN j.model
       WHEN substr(j.pin, 1, 7) = 'runner:' THEN substr(j.pin, 8) END, '')
FROM jobs j LEFT JOIN attempts a ON a.id = j.lease_attempt
WHERE j.state IN ('leased', 'running', 'draining')
