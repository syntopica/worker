INSERT INTO cooldowns (runner, until) VALUES (:runner, :until)
ON CONFLICT(runner) DO UPDATE SET until = max(until, excluded.until)
