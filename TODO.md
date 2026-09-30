# TODO

Active backlog. Closed items move to `TODO_LOG.md` with date and evidence.

## Phase 1a - engine

Closed 2026-09-30; see `TODO_LOG.md`. Live on the workstation since then.

## Phase 1b - Vexa producer (plan written when 1a lands)

- [~] Vexa producer live: worker path enabled in the installed engine
      (2026-09-30, app-data .env gets SYNTOPICA_DATA); translation proven end to
      end (40-text batch, 37 translated, rated good); production applied 23
      translation batches and 31 classifications. Open: ~60 vexa.enrich
      submissions stuck from scratch-copy key reuse, fixed in vexa a0a79e57
      but not installed (vexa main also carries another session's commits;
      vexa-8b asked to install). Ratings on ack ship with that install (148ab32c).

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

- [~] Phase 2: OpenRouter free executor, ledger and costs, quotas, `task` kind,
      retire `drip-loop.sh`. 2b (OpenRouter free rung, ledger and `costs`,
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

- [~] Weekly review on `review.weekly` writes a report (wiki eb5aef4d; a
      failed job is resubmitted along a `:r1`..`:r3` key walk). The 35B
      output is still weaker than the Claude-written one: it carries steps
      that were already done. The quality follow-up lives in the wiki
      `TODO.md`. A strong tier for `personal` would be the engine-side lever.

- [!] Routing rule 2026-09-30: codex only for adversarial review and
      consultation, cursor cancelled; bulk AI goes through the worker on agy or
      local models. Done: clips.refine runs on agy (wiki e470c8d8). Atrium lane
      agy-only: in progress (Atrium agent). Blocked: clips.grade still runs on
      codex because clips' `workerGradeTier` accepts only codex or cursor (agy
      takes no input files). Smallest unblock, in clips: grade on agy with the
      evidence inlined in the prompt, or on a local inference job.

## Known gaps accepted in the plan

- [ ] Atrium lane acks before its own registry write; a crash between them is
      recovered by the retry-key walk (`:r1`..`:r3`, atrium 0a48096) at the cost
      of one rerun. Record in 1c if observed.

## Found during phase 1a execution

- [ ] Deferred for later phases: sleep assertion for laptop nodes, remaining
      test-gap minors listed in the phase 1a final review. (Closed
      2026-09-30: split_requested timeout, fence by node, node privacy
      re-check, memory-only completion spool; status reload counter.)

## Found 2026-09-30 moving brain jobs onto the worker

- [ ] Refine and grade lost their agy fallback on a codex wall: the job now
      parks for the runner's cooldown (3600 s), and clips returns the page
      ungraded or the batch unrefined at once (`cooling_until`). Next step,
      if walls become frequent: a fallback profile per queue (e.g. agy for
      refine, which takes no inputs) chosen by the coordinator while a runner
      cools.
