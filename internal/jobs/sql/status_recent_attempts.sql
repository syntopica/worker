SELECT j.queue, a.outcome, count(*) AS n, sum(coalesce(a.wall_s, 0)) AS wall_s
FROM attempts a JOIN jobs j ON j.id = a.job_id
WHERE a.ended > :since GROUP BY j.queue, a.outcome
