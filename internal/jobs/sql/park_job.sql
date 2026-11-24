UPDATE jobs SET state = 'queued', parked = 1, parked_min_idle_s = :idle, updated = :now WHERE id = :id
