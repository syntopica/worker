# worker

Background AI job processor. `worker` runs heavy model work on your own
machines while they are idle, and only then on free or quota-limited
providers, under per-job privacy rules. Projects submit jobs through one
stable HTTP contract and collect the results without knowing where the jobs
ran.

Status: phase 1a (coordinator, node, contract v1 for inference jobs) and phase 2a (read-only `task` jobs run by codex, agy or cursor-agent). The design is in
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

For tasks, also render `com.syntopica.worker.tasks.plist.template`, substituting `@PATH@` with a PATH that reaches the runner CLIs, and declare the task profiles and the queues that grant them in the instance configuration:

```json
"profiles": {"<producer>.refine": {"runner": "codex", "model": "<model>", "env_unset": ["OPENAI_API_KEY"]}},
"queues": {"<producer>.refine": {"run_when": "active_ok", "profiles": ["<producer>.refine"]}},
"runners": {"codex": {"cooldown_s": 3600}}
```

To escalate inference to OpenRouter's free endpoints, render `com.syntopica.worker.remote.plist.template` the same way, give the node a key file that lives on the node only (`"openrouter_key_file": "~/.config/worker/openrouter.key"`, mode 600), and route a queue's local model to a `:free` model. A job escalates after waiting `after_s`; non-public classes go to zero-data-retention endpoints unless `zdr` is false, and must also be allowed the `openrouter` executor in `privacy`:

```json
"queues": {"<producer>.bulk": {"run_when": "idle", "openrouter": {"models": {"<local-model>": "<vendor>/<model>:free"}, "after_s": 600}}}
```

`worker costs --days 7` prints the ledger by day, provider and queue.

Bootstrap the agents:

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.serve.plist
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.node.plist
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.tasks.plist
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.syntopica.worker.remote.plist
```

The payload file is excluded from Time Machine automatically; exclude `$SYNTOPICA_DATA/worker/state` from any other backup tool that is not your own private copy.

Check status:

```bash
worker status
worker nodes
```

A shell script that wants one answer submits and waits in one call; `-` reads the job from stdin:

```bash
worker run --producer <producer-name> --wait 1800 job.json > answer.json
```

It prints the output (the JSON when there is one) and acknowledges it. Exit 1 means the job ended without output, 3 that it is parked behind its runner's quota wall, and 4 that it is still pending at `--wait`; in the last two cases running the same job again collects it by its idempotency key.

Both agents are LaunchAgents that require a logged-in GUI session, not LaunchDaemons.
The node reports `pressure_recovering` for its first 120 seconds, then no reason
when it may take work, or the reason it may not (`on_battery`, `user_active`,
`memory_pressure`, `backend_down`, `pressure_backoff`, `drain_failed`). The instance configuration lives in
`$SYNTOPICA_DATA/worker/config.json`; never store it in this repository. The `worker/state/`
directory must be git-ignored in the instance.

## License

MIT
