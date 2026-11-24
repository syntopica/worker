SELECT EXISTS (SELECT 1 FROM p.inputs WHERE job_id = :id),
  EXISTS (SELECT 1 FROM results r JOIN p.outputs o ON o.result_id = r.result_id WHERE r.job_id = :id)
