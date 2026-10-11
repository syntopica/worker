# TODO

Engine backlog. Closed items move to `TODO_LOG.md` with date and evidence.
This repository is public: entries describe engine behaviour only. Machines,
queue contents, live results, other repositories and owner decisions belong to
the private instance, never here.

## Phase 1c - acceptance

- [~] Cancel-to-quiet latency: read 2026-10-06 over about 6 days of real
      preemptions: 58 quiet drains, p50 4.1 s, p90 24.4 s, max 50.1 s; 8 of
      66 drains (12%) failed at 8-46 s. Drain lines carry no timestamp and do
      not split prefill from generation. Next: decide whether p90 24 s and the
      12% failed drains meet the spec, or log the phase on each drain line.
- [~] Zero model reloads while work is queued: one `worker nodes` snapshot
      (2026-10-06) showed 0 loads and 1 unload in the last hour on a macOS
      node while 17 jobs were queued, 0/0 on the Linux node. The counters
      cover a rolling hour and the nodes table keeps only the latest report,
      so a day cannot be summed afterwards. Next: persist load and unload
      events with timestamps (or sample hourly for 24 h), then count loads
      that happened while the queue was non-empty.
- [ ] Useful work under repeated interruptions: a large translation backlog
      completed with `worker status` showing progress.

## Later phases (see roadmap)

- [ ] Executor preference from judged scores: once each executor has about 20
      judged answers per queue, read `worker quality --days 7` and reorder or
      drop ladder rungs per queue. Manual for now; automating it is open.
- [~] Phase 2 remainder: a paid rung with budget reservations (the default
      budget is $0, so nothing needs it yet), and a task progress guard beyond
      the hard timeout: runners in JSON mode write stdout only at the end, so
      a guard would have to read the process group's CPU time.
- [~] Quality tiers: no example maps `tiers.strong` yet; shadow sampling
      covers the comparison meanwhile.
- [!] Phase 4: coordinator on an always-on host, Tailscale binding with
      LocalAPI WhoIs, more nodes; HID idle test under fast user switching
      before any guest node. Blocked until a candidate host has run stably
      for an agreed period.
- [~] Linux server node (33e3cb2): from the attempts table, the night of
      2026-10-05 ran clean (22 attempts, 0 preemptions), but earlier nights
      had `user_active` releases (4 on 10-01, 1 on 10-04), and no host load
      history says whether the node's own model load drove them. Next: one
      more clean night, or log the 1-minute load on each `user_active`
      release.

## Usage ledger

- [~] Task runner usage: codex, agy, cursor and max-lane record `tokens_in`
      and `tokens_out`, pinned by captured fixtures in `tests/tasks/fixtures`.
      The agy fixture is a quota-error envelope (zero counts); a success
      envelope is still uncaptured. No runner prints a price, so `cost_usd`
      stays null for every task runner. Next: capture one agy success
      envelope when its quota allows and pin its counts.

## Found during phase 1a execution

- [ ] Sleep assertion for laptop nodes; remaining minor test gaps listed in
      the phase 1a final review.
