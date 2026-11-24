SELECT j.queue, count(*) FROM attempts a JOIN jobs j ON j.id = a.job_id
WHERE a.provider = 'openrouter' AND a.started >= :midnight GROUP BY j.queue
