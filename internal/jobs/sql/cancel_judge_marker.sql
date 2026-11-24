UPDATE jobs SET state = 'cancelled', finished = :now WHERE id = :id
