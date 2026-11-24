WITH RECURSIVE queue_names(name) AS (
  SELECT min(queue) FROM jobs
  UNION ALL SELECT (SELECT min(queue) FROM jobs WHERE queue > queue_names.name)
  FROM queue_names WHERE name IS NOT NULL
)
SELECT name FROM queue_names WHERE name IS NOT NULL
