UPDATE jobs SET state = 'superseded', split_count = :count, finished = coalesce(finished, :now), updated = :now
WHERE id = :id
