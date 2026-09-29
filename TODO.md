# TODO

Active backlog. Closed items move to `TODO_LOG.md` with date and evidence.

## Phase 1a - engine (plan `docs/superpowers/plans/2026-09-29-phase-1a-engine.md`)

- [ ] Tasks 1-14: scaffold, config, store, submit, policy, leases, completion,
      results/retention, API, host state, Ollama executor, node, CLI, client and
      contract fixtures. Each task green on `uv run codeality-py gate`.
- [ ] Task 15: LaunchAgent templates; private instance config and tokens on the
      workstation.
- [ ] Task 16: Atrium local lane through the queue (`atrium` repo and the drip
      `lane.env`).
- [ ] Task 17: close phase 1a in this backlog.

## Phase 1b - Vexa producer (plan written when 1a lands)

- [ ] Rust client crate against `contract/fixtures/`.
- [ ] Translation as submit/collect with a local submission table and one
      applier; enrichment as queue `vexa.enrich` (`active_ok`); ask stays direct.

## Phase 1c - acceptance on the workstation

- [ ] Cancel-to-quiet latency on Ollama 0.34.4 (probe method, amendment 2).
- [ ] Zero model reloads across the three producers.
- [ ] Useful work under repeated interruptions; 1000 translated texts;
      `worker status` shows progress.
- [ ] Memory-pressure source: dispatch source vs sysctl (amendment 5).

## Later phases (see roadmap)

- [ ] Phase 2: OpenRouter free executor, ledger and costs, quotas, `task` kind,
      retire `drip-loop.sh`.
- [ ] Phase 3: absorb loose batch scripts from other repositories.
- [ ] Phase 4: coordinator on the always-on server, Tailscale binding, more
      nodes; HID idle test under fast user switching before any guest node.

## Known gaps accepted in the plan

- [ ] Atrium lane acks before its own registry write; a crash between them
      leaves a resubmit waiting until timeout. Record in 1c if observed.
