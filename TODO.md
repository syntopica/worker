# TODO

Engine backlog. Closed items move to `TODO_LOG.md` with date and evidence.
This repository is public: entries describe engine behaviour only. Machines,
queue contents, live results, other repositories and owner decisions belong to
the private instance, never here.

## Phase 1c - acceptance

- [~] Cancel-to-quiet latency: read 2026-10-06 over about 6 days of real
      preemptions: 58 quiet drains, p50 4.1 s, p90 24.4 s, max 50.1 s; 8 of
      66 drains (12%) failed at 8-46 s. Since 2026-10-11 each drain line
      carries `phase=`, `reason=` and a UTC `at=` (spec amendment 2026-10-11),
      effective once a node restarts on it. Next: after a few days of new
      lines, read the distribution per phase; whether p90 24 s and the failed
      share meet the spec is an owner decision.
- [~] Zero model reloads while work is queued: one `worker nodes` snapshot
      (2026-10-06) showed 0 loads and 1 unload in the last hour on a macOS
      node while 17 jobs were queued, 0/0 on the Linux node. Since 2026-10-11
      the node logs every residency change with a UTC time (spec amendment
      2026-10-11), effective once a node restarts on it. Next: after 24 h of
      lines, count the loads whose time falls while the queue held work
      (jobs' submit and finish times in the store).
- [ ] Useful work under repeated interruptions: a large translation backlog
      completed with `worker status` showing progress. Needs owner: a live
      backlog run on the instance; no engine change is pending.

## Later phases (see roadmap)

- [ ] Executor preference from judged scores: once each executor has about 20
      judged answers per queue, read `worker quality --days 7` and reorder or
      drop ladder rungs per queue. Manual for now; automating it is open.
      Needs owner: the ladder lives in the private instance configuration.
- [~] Phase 2 remainder: a paid rung with budget reservations. Needs owner:
      the default budget is $0, so nothing needs it until a paid budget is
      set. (The task progress guard is superseded, see `TODO_LOG.md`.)
- [~] Quality tiers: no example maps `tiers.strong` yet; shadow sampling
      covers the comparison meanwhile. Needs owner: which queue gets a strong
      tier is instance configuration; the engine already supports it.
- [!] Phase 4: coordinator on an always-on host, Tailscale binding with
      LocalAPI WhoIs, more nodes; HID idle test under fast user switching
      before any guest node. Blocked until a candidate host has run stably
      for an agreed period.
- [~] Linux server node (33e3cb2): from the attempts table, the night of
      2026-10-05 ran clean (22 attempts, 0 preemptions), but earlier nights
      had `user_active` releases (4 on 10-01, 1 on 10-04). Since 2026-10-11
      each drain line on Linux carries `load=` (spec amendment 2026-10-11),
      effective once the node restarts on it. Next: read the `load=` of the
      next `user_active` releases against `max_load`, or one more clean night.

## Usage ledger

- [~] Task runner usage: codex, agy, cursor and max-lane record `tokens_in`
      and `tokens_out`, pinned by captured fixtures in `tests/tasks/fixtures`.
      The agy fixture is a quota-error envelope (zero counts); a success
      envelope is still uncaptured. No runner prints a price, so `cost_usd`
      stays null for every task runner. Next: capture one agy success
      envelope when its quota allows and pin its counts.

## Found during phase 1a execution

- [!] Sleep assertion for laptop nodes; remaining minor test gaps listed in
      the phase 1a final review. Blocked: the spec names no sleep assertion,
      and holding one changes how an owner's laptop sleeps; the review's
      minor list is no longer in the repository. Smallest unblock: the
      owner decides whether a node holds an idle-sleep assertion during an
      attempt, and the gap list is recovered from the private instance.
