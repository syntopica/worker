# TODO log

Closed engine work. Public repository: engine commits and behaviour only.

## 2026

### September

- 2026-09-29: design approved and revised after three adversarial review
  rounds; spec at `docs/superpowers/specs/2026-09-29-worker-design.md`,
  phased roadmap and phase 1a plan in `docs/superpowers/plans/`.
- 2026-09-30: phase 1a engine closed (41c97f1..445dfcc, review fixes
  32190a4..d979f62). Verified: `uv run codeality-py gate` green.
- 2026-09-30 [x] Warn pressure counts only below the node's `min_free_pct`
  (default 15); critical always counts; an unreadable value still blocks.
  Verified: gate green, `tests/node/test_memory_check.py`.
- 2026-09-30 [x] Phase 1a backlog fixes (amendment "phase 1a backlog
  fixes"): loopback-only `listen`, transport backoff, failed-drain unload
  retry, per-queue candidate cap, retention validated against the unacked
  TTL, host reader failures logged by name; CI with gitleaks CLI, actionlint
  and zizmor. Verified: gate green.
- 2026-09-30 [x] Operability (amendment "operability after the first
  producers"): non-success outcomes logged with their code, `schema_path` on
  a final `schema_violation`, config reloaded on change, worker state refused
  as a task input (`input_denied`), cooldowns and `last_release` in status.
- 2026-09-30 [x] Quality tiers and model quality: `tier` on submit, executor
  model per attempt, `rating` on ack, `GET /v1/quality`, `worker quality`
  (34ae695, 468c887); `worker rate` (7c200d8).
- 2026-09-30 [x] Clean shutdown hands the attempt back as `node_shutdown`;
  a failed power-source read reuses one up to two minutes old (a907eef).
- 2026-09-30 [x] Deferred engine items (cb27f0f): fence by node, node privacy
  re-check (`privacy_refused`), unanswered `split_requested` parked after the
  unacked TTL, memory-only completion spool.
- 2026-09-30 [-] Memory-pressure dispatch source dropped (same kernel level
  as the sysctl already read); CPU-time progress guard for tasks dropped.

### October

- 2026-10-01 [x] Remote executors for every class but `secret`: credentials
  redacted from every remote prompt (e74e62b; task input files are not
  redacted); ordered OpenRouter fallback models per route (e0c4f0f).
- 2026-10-01 [x] One unreadable idle or pressure sample no longer preempts:
  a reading at most 10 s old stands in (f8f107e).
- 2026-10-01 [x] Concurrent remote slots with clean `node_shutdown` hand-back
  (1a2f13b); stop flag checked every second and one shared 12 s join so a
  restart stays inside launchd's exit timeout (7f14272).
- 2026-10-01 [x] Executor ladder: queue `runner` route, OpenRouter `after_s`
  only while that route is usable, `local_after_s`, a runner wall frees an
  inference job at once (82598f1). Adversarial review P1s fixed before
  commit.
- 2026-10-01 [x] Judged shadow sampling: sampled answers re-run on pinned
  executors and judged blind; `worker quality` prints mean score and best
  rate per queue and executor (6a22353, store v7).
- 2026-10-01 [x] Tool-less grading as inference with inlined evidence works
  on the ladder; inference jobs never park on a runner wall.
- 2026-10-01 [x] Quota walls rest `runner:model` when the profile pins a
  model, not the whole runner (9865198).
- 2026-10-01 [x] Concurrent runner task slots; the local model stays one job
  at a time (0ad1d58).
- 2026-10-01 [x] A judge gets two attempts (a6c3928); runner answers parse
  one code fence and an empty answer fails as `no_output` (6e9f506).
- 2026-10-01 [x] Headless Linux server node: `/proc` sampler with a
  `max_load` gate and per-node `coordinator_url` (33e3cb2).
- 2026-10-01 [x] A runner's empty answer to an inference job is uncharged and
  the job skips the runner rung (2c45e3a). Verified: gate green.
- 2026-10-03 [x] Status and activity: recent failures and a queue's
  `sampling_failed` leave shadow and judge jobs out (83378bd, 91ed492);
  `GET /v1/activity` aggregates finished attempts by hour with an
  `attempts(ended)` index (9b037e6). Verified: gate green, live 40 ms.
- 2026-10-03 [x] A spent quota window rests only the models it meters
  (`model_windows`, 737ee2f), so a task profile on another model family of
  the same runner can serve as its queue's fallback.
- 2026-10-03 [x] Fixed error-code vocabulary at completion, per-runner
  `max_concurrent` across nodes, and failed runners named from their stderr
  tail without keeping text (6883061). Verified: gate green.
- 2026-10-04 [x] Admin job browser, content reveal and actions (amendment
  2026-10-04, 5c0c177..3701c5c): `GET /v1/admin/jobs` (cursor-paged), job
  detail with attempts, `/content` behind `X-Worker-Reveal` for sensitive
  classes, admin cancel, retry (`retry_of`, admitted as a submit) and ack,
  an `audit` table at store version 8 and `worker audit`. Verified: gate
  green.
