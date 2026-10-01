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
- 2026-10-01 [x] Single unreadable idle/pressure sample no longer preempts:
  17 of 20 preemptions were `host_state_unreadable` under load 38; a reading
  at most 10 s old now stands in (f8f107e). Verified: gate green.
- 2026-10-01 [x] Remote loop concurrency: `remote_slots` threads, each with its
  own link and 429 rest, clean `node_shutdown` hand-back on SIGTERM (1a2f13b);
  instance set to 4 slots (wiki ed29eb49). Verified: gate green.
- 2026-10-01 [x] Executor ladder (owner order: agy first for all work, then
  OpenRouter, then local; local helps when delayed; `secret` local only):
  queue `runner` route renders inference into an agy task, OpenRouter waits
  `after_s` only while that route is usable, `local_after_s` holds local while
  a remote could take the job, an agy wall on inference frees the job at once
  (82598f1). Codex adversarial review found two P1s (task starvation behind
  the per-queue cap, node-pinned profile holding other rungs), both fixed
  before commit. Instance: `inference.agy` on all 8 inference queues,
  OpenRouter after 60 s, local after 600 s (wiki b8a8e476). Gate green.
- 2026-10-01 [x] Judged shadow sampling (owner: "comprobar la calidad de
  cada cosa para saber que usar y tener una preferencia"): sampled answers
  re-run on pinned executors, judged blind by a codex task, `worker quality`
  prints mean score and best rate per queue and executor (6a22353, store v7).
  Codex adversarial review found 3 P1 and 2 P2, all fixed with tests.
  Instance: `quality.judge` (codex gpt-5.5) on seven queues at 2-10% with
  five targets (wiki 2148bc4a). Gate green.
- 2026-10-01 [x] Atrium lane acks after its own registry write (was a known
  gap, recovered only by the retry-key walk at the cost of a rerun). Both
  lanes return the result unacked; the record carries its `result_id` and the
  ack follows the write; each pass first drains `GET /v1/results` and acks
  every result a record already holds. Atrium f72cdc9; crash-window test
  `tests/test_worker_ack_after_registry_write.py` shows one record and the
  result acked on the next pass, and fails with the old ordering.

### October

- 2026-10-01 [x] clips.grade off codex (routing rule 2026-09-30). Grade is a
  tool-less inference job on `clips.grade` with the page and its evidence
  inlined (evidence capped at 512 KB, the agy argv size measured working; the
  whole job checked against `max_payload_bytes`, 1 MiB by default), the codex
  grader's instructions and output schema unchanged. Ladder: profile
  `clips.grade` on agy pinned to gemini-3.1-pro-high, OpenRouter free models
  after 60 s, local after 600 s. clips checks the author/verifier split against
  the reported executor and discards a local answer to a prompt over its
  window. clips 2ddeb91 (pnpm check: 1264 tests), wiki 3aa85bae (config loads).
  Live: grade of `topics/contrastive-learning.md` succeeded through job
  b0da6b6c on openrouter nvidia/nemotron-3-super-120b-a12b:free (26 s, 9087
  tokens in; agy was in a quota cooldown). Closes the grade half of "Refine
  and grade lost their agy fallback": inference jobs never park on a runner
  wall.
