# TODO log

Closed engine work. Public repository: engine commits and behaviour only.

<!-- tadu:log v1 since 2026-10-12 -->

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
- 2026-10-04 [x] Admin job metrics (amendment "admin job metrics and
  running-now filter"): comma-separated `state` filter, job rows with kind,
  requested model, lease, attempt token/cost/wall totals and the newest
  attempt's model, per-attempt `wall_s` and `cost_usd`, and content-free
  `results` metadata on the detail. Verified: gate green.
- 2026-10-04 [-] OpenRouter `cost_usd` always 0 is not a bug:
  `openrouter_output` already records `usage.cost`, which OpenRouter returns
  with every response, and routes accept `:free` endpoints only, whose cost
  is 0. Revisit with the paid rung.
- 2026-10-06: `host_state_unreadable` closed by decision: accept the release
  under load (spec 7: failure = busy). About 2.8 days of kind-tagged lines held
  284 `command failed` entries, all 5 s timeouts (pmset 198, ioreg 62, sysctl
  24), none an exit status or OS error; preemptions fell from 28/day to about
  1/day and fresh timeouts coincided with a 1-minute load near 90. No longer
  timeout or bridge.
- 2026-10-07 [x] On-demand task profiles and the `max-lane` runner (spec
  amendment 2026-10-07): an `on_demand` profile is skipped by routes,
  fallbacks and pins and runs only through `worker tasks --name N --profile P`,
  which drains queued inference allowed by the profile and exits. Verified:
  `uv run codeality-py gate` green; `tests/ladder/test_on_demand_profile.py`
  and `tests/tasks/test_max_lane_runner.py`.
- 2026-10-10: a battery release now unloads the model. `relieve_battery`
  runs every node step and, while the host is on battery, unloads the
  resident models the node owns (never one resident at startup or loaded by
  another app, never a failed drain's model, which keeps its timed retry),
  one attempt per model per battery episode, so an unreachable backend cannot
  block every step and a failed unload keeps the model owned for the next one.
  `user_active` keeps the warm model as before. `probe_quiet` still renews
  `keep_alive`, which no longer matters on battery: the next step unloads.
  Three Codex review rounds. Verified: `uv run codeality-py gate` green;
  `tests/node/test_relieve_battery.py`.
- 2026-10-11 [x] Codex task usage: `codex exec` runs with `--json`; usage is
  summed over the `turn.completed` events, and failures are named from stderr
  plus the `error`/`turn.failed` event messages (spec amendment 2026-10-11).
  Fixtures captured from codex-cli 0.162.1: one success, one failed turn.
  Verified: `uv run codeality-py gate` green;
  `tests/tasks/test_codex_events.py`.
- 2026-10-11 [x] Task prompts never reach a runner's argv as options: codex
  reads the prompt on stdin (`exec -`), agy takes `--print=<prompt>` (spec
  amendment 2026-10-11). A prompt such as
  `--dangerously-bypass-approvals-and-sandbox` was parsed as codex options,
  and `resume`/`review` as subcommands (since 08cd25d). agy parsing checked
  on 1.3.3 with argv that fails before any model call. Verified:
  `uv run codeality-py gate` green; `tests/tasks/test_prompt_not_in_argv.py`.
- 2026-10-11 [~] Cancel-to-quiet split by phase: the node streams
  `/api/chat`, and each drain line names `phase=prefill|generation`, the
  release `reason=` and a UTC `at=` (spec amendment 2026-10-11). Verified:
  `uv run codeality-py gate` green; `tests/node/test_chat_stream_phase.py`
  and a chunked-encoding check of `OllamaCall`. The per-phase reading waits
  for new lines.
- 2026-10-11 [~] Linux releases name their load: `HostState.load_1m` from
  `/proc/loadavg`, appended as `load=` to the drain line. Verified:
  `uv run codeality-py gate` green; `tests/node/test_chat_stream_phase.py`,
  `tests/node/test_linux_host_sampler.py`. Reading it waits for new releases.
- 2026-10-11 [~] Residency changes persist as timestamped log lines
  (`residency load|unload model= at=`), so a day of reloads can be counted
  afterwards. Verified: `uv run codeality-py gate` green;
  `tests/node/test_track_residency.py`. The 24 h count waits for new lines.
- 2026-10-11 [~] Laptop sleep policy: a laptop node holds no idle-sleep
  assertion and hands work back through lease expiry and fencing (spec
  amendment 2026-10-11, "a laptop node holds no sleep assertion"). Decided
  2026-10-11 (owner via Codex adjudication). Observing one hand-back across a
  sleep cycle stays open in `TODO.md`.
- 2026-10-11 [~] macOS node and coordinator restarted on the phase-logging
  build (397568b, which contains 6557b5a). Decided 2026-10-11 (owner via
  Codex adjudication). No job was `leased`, `running` or `draining` before
  either restart (`worker status`, `worker nodes`, jobs table); both
  LaunchAgents were restarted with `launchctl kickstart -k`, and both run
  the editable install from this checkout. Evidence: `worker nodes` shows the
  node reporting again within 30 s, and one attempt succeeded on it after the
  restart. No preemption happened in the next 19 minutes, so no `phase=` drain
  line has been seen yet; the first one closes this.
- 2026-10-11 [-] Task progress guard beyond the hard timeout: superseded by
  spec decision "phase 1c decisions" 4 (no CPU-time progress guard for
  tasks; near-zero CPU is normal while a runner waits on a remote model, and
  `timeout_s` stays the guard).
- 2026-10-14 [x] Phase 1c acceptance closed against spec section 11.
  - Result: cancel-to-quiet latency measured per phase over 30 phase-tagged
    drains (2026-10-11..13): prefill p50 4.4 s, p90 10.4 s, max 30.3 s, no
    failure; generation p50 6.9 s, p90 18.8 s, max 74.3 s, 2 failed (7%,
    down from 12%). Model reloads: every one of 29 unloads followed a release
    (battery, user input, unreadable host state) or a quiet spell; none was a
    keep-alive expiry or an option mismatch while work waited. Useful work
    under interruptions: an instance backlog of 6694 jobs completed in seven
    days with 55 preemptions, 41 jobs finished after being preempted and none
    failed `preemption_exhausted`; `worker status` showed `done_1h` moving.
    The 1000-text translation run was replaced by that backlog because the
    translation backlog no longer exists.
  - Evidence: node.err.log `drain ... phase=` and `residency` lines; jobs and
    attempts tables; owner asked on 2026-10-14 to do every pending item.
- 2026-10-14 [x] Executor preference read from judgements: about 20 or more
  judged answers per executor and queue on the main queues. The ladder order
  already matches the scores (runner rung 3.8-4.5, OpenRouter 2.2-4.7, local
  2.0-3.2), so no rung moved.
  - Result: no configuration change.
  - Evidence: `judgements` grouped by queue, provider and model.
- 2026-10-14 [x] agy success envelope captured and pinned
  (`tests/tasks/fixtures/agy_success_envelope.json`): `input_tokens` and
  `output_tokens` map to the ledger; cached tokens are not counted.
  - Result: e612e98.
  - Evidence: `tests/tasks/test_read_usage.py` passes.
- 2026-10-14 [x] Host reader bridge widened from 10 s to 30 s: each reader
  times out after 5 s in turn, so one failed sample outlived the window and
  17 of 30 drains were still `host_state_unreadable`.
  - Result: c7c7d98, spec correction to amendment 2026-10-01.
  - Evidence: `tests/node/test_bridge_unreadable.py`; gate green.
- 2026-10-14 [x] A pinned job's empty answer is charged; it looped one shadow
  member 218 times on the same runner.
  - Result: 74dfa1c.
  - Evidence: `tests/shadow/test_pinned_no_output.py` failed before, passes.
- 2026-10-14 [x] `status` reports today's OpenRouter use, the key cap and the
  capped queues (`openrouter today N/cap capped=...`).
  - Result: 82d1f64, spec amendment 2026-10-14.
  - Evidence: `tests/jobs/test_openrouter_headroom.py`, roundtrip test; live
    status line after the coordinator restart.
- 2026-10-14 [x] A `runner_unavailable` retry waits at least ten minutes, so
  a judge no longer spends both attempts inside one runner outage.
  - Result: 7530773, spec amendment 2026-10-14.
  - Evidence: `tests/jobs/test_runner_unavailable_backoff.py`; gate green.
- 2026-10-14 [-] Phase 1a leftovers: laptop sleep and the final review's
  minor test gaps.
  - Resolution: a laptop node holds no sleep assertion (amendment
    2026-10-11); its hand-back is lease expiry plus generation fencing,
    pinned by `test_expired_lease_is_reissued_and_the_old_attempt_is_fenced`
    and `test_one_pass_reclaims_an_expired_lease`. The minor gap list is not
    recoverable: the review artifacts under `.superpowers/sdd/` were not kept
    and the named items were closed on 2026-09-30.
- 2026-10-14 [x] agy's empty answers on large prompts: a denied tool request
  ends the turn `SUCCESS` with an empty response (`denied_actions:
  RunCommand`); agy has no flag that removes its tools. Every agy prompt now
  opens with a request for a plain answer.
  - Result: 19edbce, spec amendment 2026-10-14.
  - Evidence: a 195 KB synthetic prompt answered empty in 2 of 4 runs
    without the line and in 0 of 3 with it; `test_prompt_not_in_argv.py`.
- 2026-10-14 [x] A pinned OpenRouter shadow member was sent with the queue's
  fallbacks, so a withdrawn free model was silently answered by another and
  judged under the wrong label. Leases carry `pinned`; a pinned request names
  only its own model.
  - Result: 09c5f8f, lease fixture updated, spec amendment 2026-10-14.
  - Evidence: `tests/remote/test_pinned_remote_no_fallbacks.py`; gate green.
