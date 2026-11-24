INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy,
  model, state, max_attempts, not_before, created, updated, tier, shadow_of, pin)
VALUES (:id, :producer, :queue, :kind, :key, '', 0, :privacy,
  :model, 'queued', :max_attempts, :now, :now, :now, :tier, :group, :pin);
INSERT INTO p.inputs (job_id, body) VALUES (:id, :body)
