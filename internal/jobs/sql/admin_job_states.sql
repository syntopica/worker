AND j.state IN (SELECT value FROM json_each(:states))
