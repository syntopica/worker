SELECT queue, state, count(*) AS n,
  min(CASE WHEN producer NOT IN ('_shadow', '_judge') THEN created END) AS oldest,
  sum(producer IN ('_shadow', '_judge')) AS sampling
FROM jobs GROUP BY queue, state
