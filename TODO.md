# TODO

Active backlog. Closed items move to `TODO_LOG.md` with date and evidence.

## Phase 1a - engine

Closed 2026-09-30; see `TODO_LOG.md`. Live on the workstation since then.

## Phase 1b - Vexa producer (plan written when 1a lands)

- [~] Vexa producer live: worker path enabled in the installed engine
      (2026-09-30, app-data .env gets SYNTOPICA_DATA); translation proven end to
      end (40-text batch, 37 translated, rated good); production applied 23
      translation batches and 31 classifications. Installed 2026-10-01 as build c4f35285:
      the engine closed 75 orphaned vexa.enrich submissions, acks now carry
      ratings, and store copies keep the worker path off (c228f685).

## Phase 1c - acceptance on the workstation

- [~] Cancel-to-quiet latency: the node now logs `drain quiet in <s>` for every
      preemption (amendment "phase 1c decisions", 2026-09-30). Next: read the
      distribution from node.err.log after a day of preemptions.
- [~] Zero model reloads while work is queued: keep-alive raised to 30m
      (wiki 2fe0d1dc); measure `loads_1h`/`unloads_1h` in `worker nodes`
      over a day of mixed producers.
- [ ] Useful work under repeated interruptions; 1000 translated texts;
      `worker status` shows progress. Needs Vexa's worker path enabled.

## Later phases (see roadmap)

- [ ] Executor preference from judged scores: after about a week of
      `judgements` (target 20+ judged answers per queue and executor), read
      `worker quality --days 7`, then reorder or drop rungs per queue (ladder
      `runner`/`openrouter` routes, `local_after_s`). Not automatic yet.

- [~] Phase 2: OpenRouter free executor, ledger and costs, quotas, `task` kind,
      `drip-loop.sh` retired 2026-10-01: Atrium's synthesis runs as `task`
      jobs on `atrium.tasks` (profile `atrium.agy` only, cursor dropped by the
      owner's routing rule) plus its local inference lane, produced by
      `com.cristian.atrium-synthesis` (atrium d2453c1, dotfiles b061818).
      2b (OpenRouter free rung, ledger and `costs`,
      queue fallback profiles) done 2026-09-30. 2a done 2026-09-30: read-only tasks (codex, agy,
      cursor) with profiles, workspaces, timeouts and quota-wall cooldowns
      (amendment 2026-09-30). CodexBar headroom before dispatch done
      2026-09-30 (cursor rested until its 10-Oct reset on the first probe).
      Open: paid rung with budget reservations (budget is $0 by default, so
      nothing needs it yet), and a progress guard beyond the hard timeout:
      codex, agy and cursor in JSON mode write their stdout only at the end,
      so no artifact moves mid-run; a guard would have to read the process
      group's CPU time. Write tasks and `needs_reconciliation` are no longer needed for
      clips synthesis: clips moves it through inference jobs whose page writes
      clips applies itself (`src/worker-synthesis`, another session, 2026-09-30).
- [~] Quality tiers and model quality (amendment 2026-09-30): `tier`
      (basic/strong) on submit mapped per queue, attempts record the executor
      model, `rating` on ack, `GET /v1/quality` and `worker quality` done
      2026-09-30. `worker rate` (7c200d8) lets shell producers rate after
      `worker run`; the wiki offers scan rates every answer. Open: vexa,
      atrium, clips and the weekly review do not rate yet;
      no queue maps `tiers.strong` yet; shadow sampling with a judge model is
      deferred until there is a strong executor allowed for the class.
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
      Weekly review moved: `review.weekly`, local model (wiki f5a85407).
- [!] Phase 4: coordinator on the always-on server, Tailscale binding, more
      nodes; HID idle test under fast user switching before any guest node.
      Blocked on the owner's gate (2026-09-29): the coordinator moves to the
      Mac mini only once more nodes join and the mini stops panicking. On
      2026-09-30 18:40 the mini was up 12 minutes with load 79 (a fresh
      reboot); the two M1 laptops need their users' agreement, Remote Login
      and Tailscale. Smallest unblock: a week of mini uptime without a
      watchdog panic, then a `server` node there for small models.
- [~] Night node on the hosting server (CPU only, 32 threads, ~76 GB free,
      load 1-4 through the night). 2026-10-01 probe: Ollama 0.35 in an
      isolated dir, transient systemd unit capped at CPUQuota 1200%,
      MemoryMax 40G, nice 19, idle IO. `qwen3.6:35b` warm, 12 threads:
      prompt ~120 tok/s, generation 9-11 tok/s, ~55 s for a 1.5k-token
      prompt; 16 threads under the same quota halves generation. Unit
      stopped, model kept. Next: a node there needs Tailscale (absent) and
      a night window plus host-load gate instead of the HID idle check.

- [~] Weekly review on `review.weekly` writes a report (wiki eb5aef4d; a
      failed job is resubmitted along a `:r1`..`:r3` key walk). The 35B
      output is still weaker than the Claude-written one: it carries steps
      that were already done. The quality follow-up lives in the wiki
      `TODO.md`. A strong tier for `personal` would be the engine-side lever.

- [~] Routing rule 2026-09-30: codex only for adversarial review and
      consultation, cursor cancelled; bulk AI goes through the worker on agy or
      local models. Done: clips.refine on agy (wiki e470c8d8); clips.grade on
      the executor ladder (2026-10-01, see TODO_LOG). Remaining: Atrium lane
      agy-only, in progress (Atrium agent). The only codex left on the worker
      is the `quality.judge` profile, which is review and allowed.

## Found during phase 1a execution

- [ ] Deferred for later phases: sleep assertion for laptop nodes, remaining
      test-gap minors listed in the phase 1a final review. (Closed
      2026-09-30: split_requested timeout, fence by node, node privacy
      re-check, memory-only completion spool; status reload counter.)

## Found 2026-09-30 moving brain jobs onto the worker

- [ ] Refine has no fallback on an agy wall: `clips.refine` is a task on the
      agy profile, so a wall parks it for agy's cooldown (10800 s) and clips
      keeps the first-pass verdict at once (`cooling_until`). Grade no longer
      has this problem (inference on the ladder since 2026-10-01). Not a plain
      move to the ladder: `clips.triage` already runs agy first, so refine on
      the same ladder would mostly ask the first pass's own model again. Next
      step, if walls become frequent: a `fallbacks` profile for `clips.refine`
      on a different model than triage's first rung (not codex).
