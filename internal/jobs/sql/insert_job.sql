INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy, model, state,
  max_attempts, not_before, deadline, parent_id, created, updated, tier, retry_of, exclude_providers)
VALUES (:id, :producer, :queue, :kind, :key, :digest, :priority, :privacy, :model, 'queued',
  :max_attempts, :now, :deadline, :parent_id, :now, :now, :tier, :retry_of, :exclude);
INSERT INTO p.inputs (job_id, body) VALUES (:id, :body)
