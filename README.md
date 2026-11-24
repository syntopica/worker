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

Install the Go binary from this checkout using the toolchain in `go.mod`.

```bash
CGO_ENABLED=0 GOBIN="$HOME/.local/bin" go install ./cmd/worker
export PATH="$HOME/.local/bin:$PATH"
```

For a new instance, create tokens for the coordinator, the node, and any
producers. An existing installation reuses its config and tokens. Use
`token add` only for new principals or an intentional token rotation.

```bash
worker token add --kind admin --name admin
worker token add --kind node --name '<node-name>'
worker token add --kind producer --name '<producer-name>'
```

Each token is printed once and then saved at `$SYNTOPICA_DATA/worker/state/tokens/` with mode `0600`.

On macOS, set `SYNTOPICA_DATA` to the absolute path of the instance. Set
`worker_node` to the configured node name and list the agents this machine
will run in `worker_agents`. Include `tasks` or `remote` when configured.

```bash
worker_go_bin="$HOME/.local/bin/worker"
worker_node='<node-name>'
worker_agents=(serve node)
worker_launch_dir="$HOME/Library/LaunchAgents"

xml_sed_value() {
  printf '%s' "$1" | sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/[\&#\\]/\\&/g'
}

mkdir -p "$worker_launch_dir"
for agent in "${worker_agents[@]}"; do
  plist="com.syntopica.worker.$agent.plist"
  sed -e "s#@WORKER_GO_BIN@#$(xml_sed_value "$worker_go_bin")#" \
      -e "s#@SYNTOPICA_DATA@#$(xml_sed_value "$SYNTOPICA_DATA")#" \
      -e "s#@NODE@#$(xml_sed_value "$worker_node")#" \
      -e "s#@PATH@#$(xml_sed_value "$PATH")#" \
    "launchd/$plist.template" > "$worker_launch_dir/$plist"
  plutil -lint "$worker_launch_dir/$plist"
done
```

For tasks, ensure `PATH` includes the runner CLIs when you render the
template. Declare task profiles and the queues that grant them in the
instance configuration.

```json
"profiles": {"<producer>.refine": {"runner": "codex", "model": "<model>", "env_unset": ["OPENAI_API_KEY"]}},
"queues": {"<producer>.refine": {"run_when": "active_ok", "profiles": ["<producer>.refine"]}},
"runners": {"codex": {"cooldown_s": 3600}}
```

To escalate inference to OpenRouter's free endpoints, include `remote` in
`worker_agents`, give the node a local key file (`"openrouter_key_file":
"~/.config/worker/openrouter.key"`, mode 600), and route a queue's local model
to a `:free` model. A job escalates after `after_s` seconds. Non-public classes
require the `openrouter` executor in `privacy` and use zero-data-retention
endpoints unless `zdr` is false.

```json
"queues": {"<producer>.bulk": {"run_when": "idle", "openrouter": {"models": {"<local-model>": "<vendor>/<model>:free"}, "after_s": 600}}}
```

`worker costs --days 7` prints the ledger by day, provider and queue.

For an existing installation, unload each old agent with `launchctl bootout
"gui/$(id -u)/com.syntopica.worker.<agent>"` before bootstrapping its replacement.
Bootstrap the selected agents in the current GUI session.

```bash
for agent in "${worker_agents[@]}"; do
  launchctl bootstrap "gui/$(id -u)" "$worker_launch_dir/com.syntopica.worker.$agent.plist"
done
```

The payload file is excluded from Time Machine automatically; exclude `$SYNTOPICA_DATA/worker/state` from any other backup tool that is not your own private copy.

Check status:

```bash
worker status
worker nodes
```

A shell script that wants one answer submits and waits in one call; `-` reads the job from stdin:

```bash
worker run --producer '<producer-name>' --wait 1800 job.json > answer.json
```

It prints the output (the JSON when there is one) and acknowledges it. Exit 1 means the job ended without output, 3 that it is parked behind its runner's quota wall, and 4 that it is still pending at `--wait`; in the last two cases running the same job again collects it by its idempotency key.

The LaunchAgents require a logged-in GUI session.
The node reports `pressure_recovering` for its first 120 seconds, then no reason
when it may take work, or the reason it may not (`on_battery`, `user_active`,
`memory_pressure`, `backend_down`, `pressure_backoff`, `drain_failed`). The instance configuration lives in
`$SYNTOPICA_DATA/worker/config.json`; never store it in this repository. The `worker/state/`
directory must be git-ignored in the instance.

## License

MIT
