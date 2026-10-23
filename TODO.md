# TODO

> Consolidated from the accessible agent conversation history of this
> project. Last reviewed: 2026-10-14. History coverage: Partial. Not re-audited
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

Closed 2026-10-14; see `TODO_LOG.md`.

## Later phases (see roadmap)

- [~] Phase 2 remainder: a paid rung with budget reservations. Needs owner:
      the default budget is $0, so nothing needs it until a paid budget is
      set. (The task progress guard is superseded, see `TODO_LOG.md`.)
      Decided 2026-10-11 (owner via Codex adjudication): the budget stays at $0 and the paid rung
      inactive; an explicit dollar cap comes first if paid capacity is ever
      wanted.
- [~] Quality tiers: no example maps `tiers.strong` yet. Judged shadow
      results now exist per queue (2026-10-14): the runner rung already
      scores highest wherever it answers, so no queue needs a strong tier
      yet. Needs owner: which queue gets a strong tier is instance
      configuration; the engine already supports it.
- [!] Phase 4: coordinator on an always-on host, Tailscale binding with
      LocalAPI WhoIs, more nodes; HID idle test under fast user switching
      before any guest node. Blocked until a candidate host has run stably
      for an agreed period. Decided 2026-10-11 (owner via Codex adjudication): wait; keep the
      present coordinator until a candidate host passes a stability period
      with recorded uptime, load and recovery evidence.
- [~] Linux server node (33e3cb2): each drain line on Linux carries `load=`
      (spec amendment 2026-10-11). An engine older than its configuration
      crash-loops at start (an unknown profile runner raises ValueError), so
      a server node's engine must be updated before a configuration that
      needs it is pushed. Next: read the `load=` of the next `user_active`
      releases against `max_load`, or a few clean days.

## Security

- [ ] Keep task prompts out of process arguments for the agy runner. Its current joined `--print=<prompt>` argument is visible to process inspectors; diagnostic process listings must use command names and metadata only. Next: verify the CLI's stream-json stdin contract and preserve schema support and the no-tools posture when migrating the invocation.

## Quality gate
