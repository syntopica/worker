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
