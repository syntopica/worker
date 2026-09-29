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
