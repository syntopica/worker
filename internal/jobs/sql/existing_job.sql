SELECT id, payload_hash FROM jobs WHERE producer = :producer AND queue = :queue AND idempotency_key = :key
