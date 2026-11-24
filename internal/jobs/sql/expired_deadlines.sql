SELECT id FROM jobs WHERE state = 'queued' AND deadline <= :now
