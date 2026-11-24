SELECT j.id, j.queue, j.producer, j.kind, j.model, j.priority, j.state, j.privacy, j.tier,
  j.created, j.updated, j.finished, j.deadline, j.lease_node, j.lease_expires,
  j.parent_id, j.preemptions, j.error AS last_error, j.acked, j.retry_of,
  (SELECT count(*) FROM attempts a WHERE a.job_id = j.id) AS attempts,
  (SELECT sum(a.tokens_in) FROM attempts a WHERE a.job_id = j.id) AS tokens_in,
  (SELECT sum(a.tokens_out) FROM attempts a WHERE a.job_id = j.id) AS tokens_out,
  (SELECT sum(a.cost_usd) FROM attempts a WHERE a.job_id = j.id) AS cost_usd,
  (SELECT sum(a.wall_s) FROM attempts a WHERE a.job_id = j.id) AS wall_s,
  last.model AS last_model, last.provider AS last_provider, last.started AS last_started, last.outcome AS last_outcome
FROM jobs j LEFT JOIN attempts last ON last.id = (
  SELECT a.id FROM attempts a WHERE a.job_id = j.id ORDER BY a.started DESC, a.id DESC LIMIT 1
)
