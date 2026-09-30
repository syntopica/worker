# TODO

Active backlog. Closed items move to `TODO_LOG.md` with date and evidence.

## Phase 1a - engine

Closed 2026-09-30; see `TODO_LOG.md`. Live on the workstation since then.

- [ ] Owner: exclude `$SYNTOPICA_DATA/worker/state` from Backblaze. Time
      Machine exclusion is set by the engine; Backblaze needs its own rule, and
      until then personal payloads reach that backup.

## Phase 1b - Vexa producer (plan written when 1a lands)

- [ ] Rust client crate against `contract/fixtures/`.
- [ ] Translation as submit/collect with a local submission table and one
      applier; enrichment as queue `vexa.enrich` (`active_ok`); ask stays direct.

## Phase 1c - acceptance on the workstation

- [ ] Cancel-to-quiet latency on Ollama 0.34.4 (probe method, amendment 2).
- [ ] Zero model reloads across the three producers.
      Risk seen 2026-09-29: at warn the node unloads the 22 GB model, pressure
      clears, the next lease reloads it (cold 30 GB) and warn returns - a
      reload loop. Count unloads per hour in 1c before accepting.
- [ ] Useful work under repeated interruptions; 1000 translated texts;
      `worker status` shows progress.
- [~] Memory-pressure source: dispatch source vs sysctl (amendment 5).
      2026-09-30: warn now counts only below the node's `min_free_pct`
      (amendment 2026-09-30); the dispatch-source reader is still open.
      Observed 2026-09-29 on the workstation: `kern.memorystatus_vm_pressure_level`
      read 2 (warn) with `memory_pressure` reporting 72% free, and the node
      unloaded the resident qwen3.6:35b at that moment. Measure how often warn
      fires at rest before trusting it as an unload trigger.

## Later phases (see roadmap)

- [~] Phase 2: OpenRouter free executor, ledger and costs, quotas, `task` kind,
      retire `drip-loop.sh`. 2a done 2026-09-30: read-only tasks (codex, agy,
      cursor) with profiles, workspaces, timeouts and quota-wall cooldowns
      (amendment 2026-09-30). Open: write tasks and `needs_reconciliation`
      (2b, for clips synthesis), CodexBar headroom before dispatch, OpenRouter,
      ledger and `costs`, a progress guard beyond the hard timeout, and
      `worker status` showing cooldowns.
- [ ] Phase 3: absorb loose batch scripts from other repositories.
      Started 2026-09-30 ahead of phase 2 where no `task` is needed: clips
      newsletter triage runs as inference on queue `clips.triage`
      (`runners.triage = "worker"`, clips b06252f, brain ede2b3b). Still on
      agy/codex/cursor and waiting on the `task` kind: clips synthesis, grade
      and triage refiner, and the wiki's offers scan and weekly-actions review.
- [ ] Phase 4: coordinator on the always-on server, Tailscale binding, more
      nodes; HID idle test under fast user switching before any guest node.

## Known gaps accepted in the plan

- [ ] Atrium lane acks before its own registry write; a crash between them is
      recovered by the retry-key walk (`:r1`..`:r3`, atrium 0a48096) at the cost
      of one rerun. Record in 1c if observed.

## Found during phase 1a execution

- [ ] `load_worker_config` accepts any `listen_host`, but spec says phase 1
      binds loopback only; a non-loopback value would expose the API and send
      bearer tokens in plaintext. Next step: reject non-loopback hosts in
      `load_worker_config` until phase 4's Tailscale binding.
- [ ] Client `results()` uses a fixed 40 s timeout against the server's 30 s
      `wait` cap; derive it from `wait` so a cap change cannot break it.
- [ ] If `/api/ps` answers but chat keeps failing with `transport_error`, the
      node leases again every rest and charges an attempt each time. Next step:
      back off per consecutive node-side transport failure.
- [ ] `drain_failed` clears only when the model leaves `/api/ps`; if unload and
      restart both failed, nothing retries. Next step: retry the unload on a
      slow timer while blocked.
- [ ] `load_candidates` applies `LIMIT 500` before eligibility, so more than
      500 top-priority jobs from a removed queue starve the rest.
- [ ] An exception inside `run_attempt` leaves the Ollama call running; the
      next lease queues behind it. Cancel the call in the `node_error` path.
- [ ] `release_jobs_batch` deletes unfetched `failed` control results when
      `retention_days` is shorter than `unacked_ttl`; validate
      `retention_days >= 1` and >= the unacked TTL in `parse_queue_policy`.
- [ ] Declining a split on a queue no longer in config raises inside
      `ack_result` (400 `bad_request`).
- [ ] The Time Machine exclusion failure is cached per inode and never
      retried until the file is recreated.
- [ ] `host_state_unreadable` flaps with every reader healthy. Seen
      2026-09-30 03:48: two consecutive `worker nodes` reports (`idle=None`)
      while `sample_host_state()` run by hand returned all three fields five
      times in a row, then the node recovered on its own. Next step: log which
      of ioreg/pmset/sysctl returned None (allowlisted field name only).
- [ ] Deferred for later phases: `split_requested` timeout, `check_fence`
      node match and lease privacy re-check (before guest nodes), completion
      spool (spec 10), status gaps (spec 13, reload counter needed by 1c),
      sleep assertion for laptop nodes, remaining test-gap minors listed in
      the phase 1a final review.
