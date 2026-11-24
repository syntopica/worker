SELECT j.queue queue, j.tier tier, coalesce(a.provider, 'unknown') provider, coalesce(a.model, j.model) model, count(*) attempts, sum(a.outcome='succeeded') succeeded, sum(a.outcome='schema_violation') schema_violations, sum(a.outcome='failed' AND coalesce(a.error, '') NOT IN ('rate_limited', 'quota_wall')) failed, sum(a.outcome='preempted') preempted, avg(a.wall_s) mean_wall_s
FROM attempts a JOIN jobs j ON j.id=a.job_id
WHERE a.started>:since AND a.ended IS NOT NULL
GROUP BY 1, 2, 3, 4
ORDER BY 1, 2, 3, 4
