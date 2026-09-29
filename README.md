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

## Running it

Install the `worker` command:

```bash
uv tool install --editable .
```

Create tokens for the coordinator, the node, and any producers:

```bash
worker token add --kind admin --name admin
worker token add --kind node --name <node-name>
worker token add --kind producer --name <producer-name>
```

Each token is printed once and then saved at `$SYNTOPICA_DATA/worker/state/tokens/` with mode `0600`.

Render the LaunchAgent templates by substituting `@WORKER_BIN@`, `@SYNTOPICA_DATA@`, and `@NODE@`:

```bash
for t in serve node; do
  sed -e "s#@WORKER_BIN@#$(command -v worker)#" \
      -e "s#@SYNTOPICA_DATA@#$SYNTOPICA_DATA#" \
      -e "s#@NODE@#<node-name>#" \
    launchd/com.syntopica.worker.$t.plist.template > \
    ~/Library/LaunchAgents/com.syntopica.worker.$t.plist
done
```

Bootstrap both agents:

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.serve.plist
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.node.plist
```

Check status:

```bash
worker status
worker nodes
```

Both agents are LaunchAgents that require a logged-in GUI session, not LaunchDaemons.
The node reports `pressure_recovering` for its first 120 seconds, then no reason
when it may take work, or the reason it may not (`on_battery`, `user_active`,
`memory_pressure`, `backend_down`, `pressure_backoff`, `drain_failed`). The instance configuration lives in
`$SYNTOPICA_DATA/worker/config.json`; never store it in this repository. The `worker/state/`
directory must be git-ignored in the instance.

## License

MIT
