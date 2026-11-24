UPDATE jobs SET state = 'queued', parked_runs = :runs, parked_min_idle_s = parked_min_idle_s * 2,
  updated = :now, lease_attempt = NULL, lease_expires = NULL
WHERE id = :id
