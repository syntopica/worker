# TODO log

## 2026

### September

- 2026-09-29: design approved with the owner and revised after two adversarial
  review rounds; spec at `docs/superpowers/specs/2026-09-29-worker-design.md`.
- 2026-09-29: third Codex round closed the last blockers (split admission at
  the outstanding limit, bounded parked retries with `preemption_exhausted`).
  Loop stopped at the declared cap of three rounds.
- 2026-09-29: phased roadmap and detailed phase 1a plan written
  (`docs/superpowers/plans/`); spec amendments appended for the plan's
  decisions (control results for failed/expired, probe-based quiet detection,
  instance layout, producer grants, pressure source).
- 2026-09-30: phase 1a engine closed. Tasks 1-16 executed task by task with a
  review after each (41c97f1..445dfcc); final whole-branch review found one
  critical (quadratic retention sweep under the write lock) and eight
  important defects, fixed in 32190a4..d979f62 and re-reviewed. Gate
  `uv run pytest -q` (255 passed) and `uv run codeality-py gate` green at
  d979f62. Live on the workstation: LaunchAgents
  `com.syntopica.worker.serve` and `.node`, store at version 3 with payload
  files 0600 and excluded from Time Machine; Atrium drip lane `worker`
  submitting to `atrium.synthesis` (atrium cc1727d, 6b3a1b6, 0a48096;
  dotfiles 60da964).

- 2026-09-30 [x] Warn pressure no longer unloads the node's own model with a
  third of memory free. `kern.memorystatus_vm_pressure_level` read 2 with
  28-30% free (`kern.memorystatus_level`), preempting the first `clips.triage`
  job twice in 45 min. Warn now counts only below the node's `min_free_pct`
  (default 15); critical always counts; an unreadable value still blocks.
  Spec amendment 2026-09-30. Verified: `uv run codeality-py gate` (all
  passed), `tests/node/test_memory_check.py`.

- 2026-09-30 [-] Backblaze exclusion of the worker state: owner decision, the
  Backblaze backup is the owner's own private copy, so payloads there are
  accepted.
- 2026-09-30 [x] Ten phase 1a engine defects closed (spec amendment
  "phase 1a backlog fixes"): loopback-only `listen`, `results()` timeout
  derived from `wait`, transport-failure backoff, failed-drain unload retry,
  per-queue candidate cap that ignores removed queues, cancel of the backend
  call on an attempt exception, `retention_days` validated against the unacked
  TTL, split decline on a removed queue, Time Machine exclusion retried after a
  failure, host reader failures logged by name. CI fixed as well: gitleaks CLI
  instead of the licensed action, actionlint and zizmor added, mypy no longer
  host-dependent. Verified: `uv run codeality-py gate` all passed.
- 2026-09-30 [x] Operability items found moving brain jobs onto the worker
  (spec amendment "operability after the first producers"): non-success
  attempt outcomes logged with their code, `schema_path` on a final
  `schema_violation`, `config.json` reloaded on change by serve, node and
  tasks, worker state refused as a task input (`input_denied`, closes the
  `clips.grade` broad `input_root` exposure of tokens), `cooldowns` and
  per-node `last_release` in status. Verified: `uv run codeality-py gate`.
- 2026-09-30 [x] Quality tiers and model quality: `tier` on submit, executor
  model per attempt, `rating` on ack, `GET /v1/quality` and `worker quality`
  (34ae695, 468c887). Verified: gate green, live report after restart.
- 2026-09-30 [x] Clean shutdown and pmset hold (a907eef): SIGTERM hands the
  running attempt back as `node_shutdown` (no attempt or preemption charged);
  a failed pmset read reuses a reading up to two minutes old. Cause of the
  weekly review's `lease_lost`: service restarts. Verified: gate green on a
  clean export of HEAD.
- 2026-09-30 [x] `worker rate` and the result id on `worker run` stderr
  (7c200d8), so a shell producer can rate an output after judging it.
  Verified: gate green (373 passed), test `test_run_names_the_result_so_the_script_can_rate_it`.
- 2026-09-30 [x] Wiki offers scan on `offers.classify` (wiki eb5aef4d): notes
  are kept only from chunks with offers, and short references replace
  Message-IDs. The prefilter passed with string references (47 of 47
  matched). The earlier 2 schema violations were the old `["string","number"]`
  prefilter schema. Answers are rated good or edited. Verified by four 7-day
  runs on 2026-09-30.
- 2026-09-30 [x] Deferred engine items (cb27f0f): fence by node, node privacy
  re-check (`privacy_refused`), unanswered `split_requested` parked after the
  unacked TTL, memory-only completion spool. Verified: gate green.
- 2026-09-30 [-] Memory-pressure dispatch source: dropped, same kernel level
  as the sysctl already read (amendment "phase 1c decisions"). CPU-time
  progress guard for tasks: dropped, runners idle on network by design.
- 2026-10-01 [x] OpenRouter credit: owner bought $10; `/api/v1/key` now reports
  `is_free_tier: false` and free-model daily requests limit 1000 (was 50).
  Credits may expire 365 days after purchase (OpenRouter terms).
- 2026-10-01 [x] Remote for every class but `secret` (owner: "no me importa
  mandar cosas mientras no mandemos passwords"): credentials redacted from
  every remote prompt (e74e62b; task input files are not redacted), `mail` and
  `internal` opened to runner and OpenRouter, OpenRouter routes on
  vexa.enrich/translate, clips.triage/synthesis, offers.classify,
  agents.classify, atrium.synthesis (wiki 7231b04c). Ordered free fallback
  models per route (e0c4f0f, wiki e14de422) after ~50% of calls to one free
  model hit 429. Verified: gate green, services restarted.
