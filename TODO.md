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

## Node

- [~] `host_state_unreadable` on a macOS node: about 2.8 days of kind-tagged
      lines held 284 `command failed` entries, all 5 s timeouts (pmset 198,
      ioreg 62, sysctl 24), none an exit status or OS error. Preemptions from
      it fell from 28/day to about 1/day, and fresh timeouts coincided with a
      1-minute load near 90. Recommendation: accept the release under load
      (spec 7: failure = busy) rather than lengthen the timeout or bridge.
      Next: owner confirms, then close.

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

- [~] Task runners record no tokens and no cost. agy and cursor are now
      mapped: `read_usage` takes `tokens_in`/`tokens_out` from the envelope's
      `usage` (agy `input_tokens`/`output_tokens`, cursor
      `inputTokens`/`outputTokens`), pinned by captured fixtures in
      `tests/tasks/fixtures`. The agy fixture is a quota-error envelope (zero
      counts); a success envelope is still uncaptured. No envelope carries a
      price, so `cost_usd` stays null for every task runner. Codex is still
      unmeasured: without `--json` it prints only a locale-formatted total on
      stderr; with `--json` its `turn.completed` event carries `usage`
      (`input_tokens`, `cached_input_tokens`, `output_tokens`,
      `reasoning_output_tokens`), but its error text then moves from stderr to
      stdout `error`/`turn.failed` events, which `runner_failure_code` (stderr
      only) would stop seeing. Next: add `--json` to `codex_invocation`, feed
      the `turn.failed` message to `runner_failure_code`, and parse the last
      `turn.completed` usage, with a captured JSONL fixture.

## Found during phase 1a execution

- [ ] Sleep assertion for laptop nodes; remaining minor test gaps listed in
      the phase 1a final review.
