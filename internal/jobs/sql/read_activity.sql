SELECT cast(a.ended / :bucket AS INTEGER) * :bucket bucket, j.queue queue, coalesce(a.provider, 'unknown') provider, j.producer IN ('_shadow', '_judge') sampling, a.outcome outcome, CASE WHEN a.outcome = 'succeeded' THEN NULL ELSE a.error END error, count(*) attempts, sum(coalesce(a.wall_s, 0)) wall_s, sum(coalesce(a.tokens_in, 0)) tokens_in, sum(coalesce(a.tokens_out, 0)) tokens_out
FROM attempts a JOIN jobs j ON j.id = a.job_id
WHERE a.ended > :since AND a.ended <= :now
GROUP BY 1, 2, 3, 4, 5, 6
ORDER BY 1, 2, 3, 4, 5, 6
