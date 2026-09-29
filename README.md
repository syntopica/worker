# worker

Background AI job processor. `worker` runs heavy model work on your own
machines while they are idle, and only then on free or quota-limited
providers, under per-job privacy rules. Projects submit jobs through one
stable HTTP contract and collect the results without knowing where the jobs
ran.

Status: design approved, no code yet. The design is in
[`docs/superpowers/specs/2026-09-29-worker-design.md`](docs/superpowers/specs/2026-09-29-worker-design.md).

## Shape

- A **coordinator** stores, authorizes, schedules and accounts for jobs. It
  keeps them in SQLite and executes nothing.
- A **node** runs on each machine. It takes work only while that machine is
  idle, on AC power and without memory pressure, and hands the work back the
  moment someone uses the machine again. The node is the only managed caller
  of the machine's local model server, so the model is never reloaded because
  two clients disagree on its options.
- Jobs come in two kinds:
  - `inference`: a portable model call;
  - `task`: an authorized runner over an authorized input set.
- Routing follows a cost ladder: a local model first, then free endpoints,
  then subscription headroom, and paid endpoints only within a budget. A
  job's privacy class can cut the ladder short. Mail, for example, stays
  local by default.

## Instance data

This repository holds the engine only. Machines, tokens, provider allowlists,
privacy policy, queues, the database and the logs all live in a private data
instance, located through `SYNTOPICA_DATA` or the enclosing
`syntopica.config.json`, the same way as the other Syntopica engines.

## License

MIT
