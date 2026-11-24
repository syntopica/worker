SELECT producer, queue, state, coalesce(split_count, 0), idempotency_key FROM jobs WHERE id = :id
