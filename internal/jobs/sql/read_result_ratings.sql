SELECT r.queue queue, coalesce(j.tier, 'basic') tier, coalesce(json_extract(r.executor, '$.provider'), 'unknown') provider, coalesce(nullif(json_extract(r.executor, '$.model'), ''), j.model, 'unknown') model, count(*) results, sum(r.rating IS NOT NULL) rated, sum(r.rating IS 'good') good, sum(r.rating IS 'edited') edited, sum(r.rating IS 'discarded') discarded
FROM results r LEFT JOIN jobs j ON j.id=r.job_id
WHERE r.created>:since AND r.control IS NULL AND r.producer NOT IN ('_shadow', '_judge')
GROUP BY 1, 2, 3, 4
ORDER BY 1, 2, 3, 4
