UPDATE attempts SET ended = :now, outcome = :outcome, error = :code, wall_s = :wall_s,
  tokens_in = :tokens_in, tokens_out = :tokens_out, provider = :provider, cost_usd = :cost_usd,
  model = :model
WHERE id = :attempt
