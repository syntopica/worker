SELECT coalesce(a.provider, 'unknown') provider, j.queue queue, date(a.ended, 'unixepoch') day, count(*) attempts, sum(a.outcome='succeeded') succeeded, sum(coalesce(a.tokens_in, 0)) tokens_in, sum(coalesce(a.tokens_out, 0)) tokens_out, sum(coalesce(a.cost_usd, 0)) cost_usd, sum(coalesce(a.wall_s, 0)) wall_s
FROM attempts a JOIN jobs j ON j.id=a.job_id
WHERE a.started>:since AND a.ended IS NOT NULL
GROUP BY provider, queue, day
ORDER BY day, provider, queue
