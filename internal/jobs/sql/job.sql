SELECT id, producer, queue, kind, idempotency_key, priority, privacy, model, state, attempts, max_attempts,
  preemptions, parked, parked_runs, parked_min_idle_s, generation, coalesce(lease_node, ''),
  coalesce(lease_attempt, ''), coalesce(error, ''), tier, coalesce(shadow_of, ''), coalesce(pin, ''),
  coalesce(exclude_providers, '')
FROM jobs WHERE id = :id
