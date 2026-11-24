SELECT id, state FROM jobs WHERE lease_attempt = :attempt AND generation = :generation AND lease_node = :node
