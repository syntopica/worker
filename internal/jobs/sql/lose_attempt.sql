UPDATE attempts SET ended = :now, outcome = 'lost' WHERE id = :attempt;
UPDATE jobs SET attempts = attempts + 1 WHERE id = :id
