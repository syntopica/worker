SELECT j.queue, count(*) FROM attempts a JOIN jobs j ON j.id = a.job_id
WHERE a.started > :since GROUP BY j.queue
