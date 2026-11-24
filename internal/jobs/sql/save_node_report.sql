INSERT INTO nodes (name, report, updated) VALUES (:node, :report, :now)
ON CONFLICT(name) DO UPDATE SET report = excluded.report, updated = excluded.updated
