UPDATE jobs SET state = 'succeeded', finished = :now, attempts = 1 WHERE id = :id
