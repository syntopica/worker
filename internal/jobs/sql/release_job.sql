DELETE FROM results WHERE job_id = :id;
DELETE FROM attempts WHERE job_id = :id;
DELETE FROM jobs WHERE id = :id
