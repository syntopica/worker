UPDATE results SET acked = coalesce(acked, :now), rating = coalesce(:rating, rating)
WHERE result_id = :result_id
