SELECT created AS time, action, job_id, privacy AS class, principal
FROM audit WHERE created > :since ORDER BY seq DESC
