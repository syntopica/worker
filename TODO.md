# TODO

Engine backlog. Closed items move to `TODO_LOG.md` with date and evidence.
This repository is public: entries describe engine behaviour only. Machines,
queue contents, live results, other repositories and owner decisions belong to
the private instance, never here.

## Phase 1c - acceptance

- [~] Cancel-to-quiet latency: the node logs `drain quiet in <s>` for every
      preemption (amendment "phase 1c decisions"). Next: read the
      distribution after a day of preemptions.
- [~] Zero model reloads while work is queued: measure `loads_1h` and
      `unloads_1h` in `worker nodes` over a day of mixed producers.
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
- [~] Linux server node (33e3cb2): confirm over a full night window that the
      node's own model load does not trip `max_load` and release idle-only
      work as `user_active`.

## Found during phase 1a execution

- [ ] Sleep assertion for laptop nodes; remaining minor test gaps listed in
      the phase 1a final review.
