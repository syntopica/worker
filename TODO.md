# TODO

> Consolidated from the accessible agent conversation history of this
> project. Last reviewed: 2026-10-12. History coverage: Partial. Not re-audited
> during the 2026-10-12 format migration.
>
> States: `[ ]` pending · `[~]` partial or unverified · `[!]` blocked · `[x]`
> verified complete · `[-]` obsolete or superseded. Closed work moves to
> `TODO_LOG.md`.

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
      share meet the spec is an owner decision. Decided 2026-10-11 (owner via Codex adjudication): wait;
      do not accept 1c from the aggregate p90, read prefill and generation
      and the failed drains after several days of phase-tagged lines. A node
      restarted on this build on 2026-10-11.
- [~] Zero model reloads while work is queued: one `worker nodes` snapshot
      (2026-10-06) showed 0 loads and 1 unload in the last hour on a macOS
      node while 17 jobs were queued, 0/0 on the Linux node. Since 2026-10-11
      the node logs every residency change with a UTC time (spec amendment
      2026-10-11), effective once a node restarts on it (a macOS node did on
      2026-10-11). Next: after 24 h of lines, count the loads whose time falls while the queue held work
      (jobs' submit and finish times in the store).
- [ ] Useful work under repeated interruptions: a large translation backlog
      completed with `worker status` showing progress. Needs owner: a live
      backlog run on the instance; no engine change is pending. Decided 2026-10-11 (owner via Codex adjudication): 
      run it under ordinary interruptions and require completed work plus
      visible progress. Still needs owner: naming which instance backlog is
      authorized for the live run.

## Later phases (see roadmap)

- [ ] Executor preference from judged scores: once each executor has about 20
      judged answers per queue, read `worker quality --days 7` and reorder or
      drop ladder rungs per queue. Manual for now; automating it is open.
      Needs owner: the ladder lives in the private instance configuration.
      Decided 2026-10-11 (owner via Codex adjudication): wait; keep the current ladder until the
      threshold is met.
- [~] Phase 2 remainder: a paid rung with budget reservations. Needs owner:
      the default budget is $0, so nothing needs it until a paid budget is
      set. (The task progress guard is superseded, see `TODO_LOG.md`.)
      Decided 2026-10-11 (owner via Codex adjudication): the budget stays at $0 and the paid rung
      inactive; an explicit dollar cap comes first if paid capacity is ever
      wanted.
- [~] Quality tiers: no example maps `tiers.strong` yet; shadow sampling
      covers the comparison meanwhile. Needs owner: which queue gets a strong
      tier is instance configuration; the engine already supports it.
      Decided 2026-10-11 (owner via Codex adjudication): wait; map no queue to `tiers.strong`
      until judged queue-level shadow results exist.
- [!] Phase 4: coordinator on an always-on host, Tailscale binding with
      LocalAPI WhoIs, more nodes; HID idle test under fast user switching
      before any guest node. Blocked until a candidate host has run stably
      for an agreed period. Decided 2026-10-11 (owner via Codex adjudication): wait; keep the
      present coordinator until a candidate host passes a stability period
      with recorded uptime, load and recovery evidence.
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

- [~] Sleep assertion for laptop nodes; remaining minor test gaps listed in
      the phase 1a final review. Decided 2026-10-11 (owner via Codex adjudication): a laptop
      node holds no idle-sleep assertion; a sleeping laptop hands its work
      back through lease expiry and fencing (spec amendment 2026-10-11, "a
      laptop node holds no sleep assertion"). Remaining: observe one hand-back
      across a normal sleep cycle, and recover the review's minor gap list
      from the private instance.
