SELECT runner, until - :now AS remaining FROM cooldowns WHERE until > :now
