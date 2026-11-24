SELECT node, error, :now - max(ended) AS age_s FROM attempts
WHERE started > :since AND outcome = 'preempted' GROUP BY node
