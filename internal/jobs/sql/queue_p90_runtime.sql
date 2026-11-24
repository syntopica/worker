WITH recent AS (
  SELECT a.wall_s FROM attempts a JOIN jobs j ON j.id = a.job_id
  WHERE j.queue = :queue AND a.outcome = 'succeeded' AND a.wall_s IS NOT NULL
  ORDER BY a.started DESC LIMIT 200
)
SELECT wall_s FROM recent ORDER BY wall_s LIMIT 1
OFFSET (SELECT count(*) * 9 / 10 FROM recent)
