SELECT id FROM jobs WHERE state IN ('leased', 'running', 'draining') AND lease_expires < :now
