SELECT id FROM jobs WHERE state IN ('queued', 'leased', 'running', 'draining', 'split_requested')
  AND id NOT IN (SELECT job_id FROM p.inputs)
