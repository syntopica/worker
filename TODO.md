# TODO

Active backlog. Closed items move to `TODO_LOG.md` with date and evidence.

## Phase 1a - engine

Closed 2026-09-30; see `TODO_LOG.md`. Live on the workstation since then.

## Phase 1b - Vexa producer (plan written when 1a lands)

- [ ] Rust client crate against `contract/fixtures/`.
- [ ] Translation as submit/collect with a local submission table and one
      applier; enrichment as queue `vexa.enrich` (`active_ok`); ask stays direct.

## Phase 1c - acceptance on the workstation

- [ ] Cancel-to-quiet latency on Ollama 0.34.4 (probe method, amendment 2).
- [ ] Zero model reloads across the three producers. Measure with
      `loads_1h`/`unloads_1h` in `worker nodes` (2026-09-30). Note the pinned
      `keep_alive` of 5m unloads the model after any 5-minute gap, which alone
      breaks this criterion between bursts; decide the keep-alive with it.
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
      retire `drip-loop.sh`. 2b (OpenRouter free rung, ledger and `costs`,
      queue fallback profiles) done 2026-09-30. 2a done 2026-09-30: read-only tasks (codex, agy,
      cursor) with profiles, workspaces, timeouts and quota-wall cooldowns
      (amendment 2026-09-30). Open: CodexBar headroom before dispatch, paid
      rung with budget reservations, and a progress guard beyond the hard
      timeout. Write tasks and `needs_reconciliation` are no longer needed for
      clips synthesis: clips moves it through inference jobs whose page writes
      clips applies itself (`src/worker-synthesis`, another session, 2026-09-30).
- [ ] Phase 3: absorb loose batch scripts from other repositories.
      Started 2026-09-30 ahead of phase 2 where no `task` is needed: clips
      newsletter triage runs as inference on queue `clips.triage`
      (`runners.triage = "worker"`, clips b06252f, brain ede2b3b). Also moved
      2026-09-30: triage refiner (`clips.refine` task, clips c643ae7), grade
      (`clips.grade` task with an evidence manifest, clips ab681f3) and the
      wiki's offers scan (`offers.classify`, `mail` inference on the local
      model, via `worker run`). Left: clips synthesis (in progress in clips
      through inference, another session), the weekly-actions review (owner
      chose 2026-09-30 a local-model rewrite over a precomputed manifest; in
      progress). Done 2026-09-30: agents library-loop classify stage on queue
      `agents.classify` (wiki `tools/agents/worker-classifier.sh`; its old agy
      classifier path had been failing every run), and queue `clips.synthesis`
      declared for the clips engine. Transcription: no scheduled job exists
      (only an unscheduled prototype), nothing to absorb.
- [!] Phase 4: coordinator on the always-on server, Tailscale binding, more
      nodes; HID idle test under fast user switching before any guest node.
      Blocked on the owner's gate (2026-09-29): the coordinator moves to the
      Mac mini only once more nodes join and the mini stops panicking. On
      2026-09-30 18:40 the mini was up 12 minutes with load 79 (a fresh
      reboot); the two M1 laptops need their users' agreement, Remote Login
      and Tailscale. Smallest unblock: a week of mini uptime without a
      watchdog panic, then a `server` node there for small models.

## Known gaps accepted in the plan

- [ ] Atrium lane acks before its own registry write; a crash between them is
      recovered by the retry-key walk (`:r1`..`:r3`, atrium 0a48096) at the cost
      of one rerun. Record in 1c if observed.

## Found during phase 1a execution

- [ ] Deferred for later phases: `split_requested` timeout, `check_fence`
      node match and lease privacy re-check (before guest nodes), completion
      spool (spec 10), status gaps (spec 13; reload counter done 2026-09-30),
      sleep assertion for laptop nodes, remaining test-gap minors listed in
      the phase 1a final review.

## Found 2026-09-30 moving brain jobs onto the worker

- [ ] Refine and grade lost their agy fallback on a codex wall: the job now
      parks for the runner's cooldown (3600 s), and clips returns the page
      ungraded or the batch unrefined at once (`cooling_until`). Next step,
      if walls become frequent: a fallback profile per queue (e.g. agy for
      refine, which takes no inputs) chosen by the coordinator while a runner
      cools.
- [ ] Offers scan (wiki `tools/offers/scan.sh`) on the local model: 7-day run
      exit 0 in 22 min, 9 offers, but the 40k-token context forces 40-body
      chunks, and chunk `notes` are joined verbatim, so one chunk's "no audio
      offers" contradicts another's XLN sale. Next steps: verify the prefilter
      passes with string IDs on the next run; merge notes by keeping only
      notes from chunks that returned offers, or summarise them in one extra
      small job.
