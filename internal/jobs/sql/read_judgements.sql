SELECT queue, provider, model, count(*) judged, round(avg(score), 2) mean_score, sum(best) best
FROM judgements
WHERE created>:since
GROUP BY 1, 2, 3
ORDER BY 1, 5 DESC
