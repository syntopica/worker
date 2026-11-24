SELECT max(until) FROM cooldowns WHERE runner IN (:runner, :model_key) AND until > :now
