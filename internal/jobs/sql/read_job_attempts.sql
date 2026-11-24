SELECT node, provider, model, outcome, error, started, ended, tokens_in, tokens_out, wall_s, cost_usd
FROM attempts
WHERE job_id=:id
ORDER BY started, id
