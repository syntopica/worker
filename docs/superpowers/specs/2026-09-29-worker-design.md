# worker - design

Status: approved design, revised after an adversarial review. Date: 2026-09-29.

## 1. Purpose

`worker` is a background processor for heavy AI work. It runs that work on the
owner's own machines while they are idle, and only then on free or
quota-limited providers, under per-job privacy rules. Producers submit jobs
through one stable contract and collect results without knowing where a job
ran.

It replaces the uncoordinated pattern where every project runs its own loop
against the same GPU and the same model: idle checks copied between projects,
drip loops, codex and agy batch passes, and direct Ollama callers that force
model reloads on each other. It is a job processor, not an assistant, and it is
meant to be reusable by any project, not only the ones that motivated it.

This repository is public and holds the engine only. Everything
instance-specific lives in the private data instance, located the same way as
the other Syntopica engines (`SYNTOPICA_DATA`, or the enclosing
`syntopica.config.json` with `syntopica.local.json` taking precedence). That
covers machines, addresses, tokens, provider allowlists, privacy policy, queue
definitions, the database and the logs.

## 2. Workloads that motivated it

These are the reference producers. The engine must not hard-code any of them.

| Producer | Work | Today |
| --- | --- | --- |
| Vexa enrichment | Mail classification, JSON output, one message per call | Ollama `qwen3.6:35b` with `num_ctx` 40960. Runs whenever mail arrives. Serialized inside Vexa by resource `ai:hosted`. |
| Vexa translation | A large backlog of already-extracted texts, sent in batches of 40, with a check that digit sequences match the source | Same model. Tens of seconds per warm batch. Runs only after 5 min idle, on AC power. |
| Atrium synthesis drip | A large episode backlog | Lanes are cursor, agy and local Ollama, the local lane idle-only. Quotas are read through CodexBar. Has a wall cooldown and a stall guard. |
| Atrium refresh | ONNX embeddings | CPU, in-process, hourly. |
| brain and clips | Bulk reading, screening and classification | `codex exec`, and agy/Gemini with schema output. Sensitive pages must never reach cloud runners. |
| Others | Periodic classification loops, transcription, scrapers with LLM triage | Various CLIs, cron or launchd. |

Measured constraint: Ollama reloads a loaded model whenever the normalized
runner options of a request differ from the loaded runner, and also when
adapters, projectors, the context-shift setting or runner health change. With
a 22 GB model a reload costs tens of seconds and kills in-flight requests.

## 3. Reference hardware

The engine reads machines from instance configuration; the actual fleet is
documented in the private instance. The design has to cover:

- a large-memory workstation that is the only machine able to fit the large
  model, and that sleeps, moves and runs on battery;
- small-memory always-on machines that already host other services;
- machines used by other people, which can only offer spare time as guests.

## 4. Decisions

1. **One engine owns all AI job execution.** Assistant and session control stay
   in other projects.
2. **Two job kinds.**
   - `inference` is a portable model call: messages, an optional JSON schema,
     and model requirements. Any compatible executor can run it.
   - `task` runs an authorized runner (codex, agy, cursor-agent) over an
     authorized input set, on a machine that holds that runner's credentials.
3. **A custom coordinator.**
   - It is Python with uv, over SQLite in WAL mode, and exposes an HTTP API.
   - There is one pull-based node per machine.
   - Rejected alternatives:
     - Postgres with a queue library. Idle gating, model affinity, cost routing
       and quotas are custom anyway.
     - Temporal or Hatchet. Heavy, and a foreign programming model.
4. **The coordinator's home follows the fleet.**
   - In phase 1, coordinator and node both run on the workstation, bound to
     loopback. That is where the only capable node and the producers are, so a
     remote control plane would add failure modes without adding capacity.
   - It moves to the always-on server once more nodes join and that host has
     shown it is stable (section 12). The contract does not change.
5. **Idle-only by default.**
   - A queue may declare `run_when: active_ok`, with a concurrency of 1, for
     low-volume, latency-sensitive work such as new-mail enrichment.
6. **At-least-once execution and delivery, made safe by fencing and by
   producer-side deduplication.** Exactly-once delivery is not claimed.

## 5. Components

```
producers -- HTTP/JSON (loopback in phase 1, Tailscale later), per-producer token
    |
    v
coordinator -- SQLite WAL (in the private instance)
    - API: submit / get / list / wait / cancel / ack
    - scheduler: grants a lease with an attempt id and a fencing generation
    - policy: authorization, privacy, budgets (read from instance config)
    - ledger: measured usage per attempt
    ^
    | pull lease / heartbeat / complete (fenced)
node (one per machine, user LaunchAgent)
    - host state: HID idle, power source, memory pressure
    - executors: ollama | openrouter | runner (task) | local-cpu
    - sole Ollama client on its machine for the managed models
CLI: status | nodes | queue | jobs | job | cancel | retry | submit | costs
```

- The coordinator executes nothing. It stores, authorizes, schedules and
  accounts.
- The node is the only managed caller of Ollama on its machine.
  - Reload-sensitive runner options are pinned per model in instance
    configuration: `num_ctx`, `num_batch`, `OLLAMA_NUM_PARALLEL`,
    `OLLAMA_MAX_LOADED_MODELS` and the keep-alive.
  - Per-request sampling options (temperature, format) pass through.
  - Exclusivity is a cooperative convention between enrolled applications, not
    OS-level isolation. Local Ollama has no authentication. The node reports
    any unexpected loaded runner it sees in `/api/ps` and any reload it
    observes, so a violation is visible rather than silent.
- Ollama's cloud forwarding is disabled on every node that runs local-only
  classes.
- Documented exception: Vexa's interactive ask path stays a direct caller. It
  uses the same pinned options, and its volume is negligible. It is listed in
  instance configuration so the node knows the exception exists.

## 6. Job contract v1

`POST /v1/jobs`

```json
{
  "contract": 1,
  "kind": "inference",
  "queue": "vexa.translate",
  "idempotency_key": "tr:<batch-id>",
  "priority": 50,
  "privacy": "mail",
  "deadline": null,
  "max_attempts": 3,
  "requirements": {
    "capability": "chat.json",
    "models": ["qwen3.6:35b"],
    "min_context": 32000
  },
  "input": {"messages": [], "schema": {}, "options": {"temperature": 0}}
}
```

A `task` input carries:

```json
{
  "runner": "codex",
  "profile": "<authorized profile>",
  "inputs": ["<manifest entries>"],
  "prompt": "...",
  "output_schema": {},
  "sandbox": "read-only"
}
```

The coordinator resolves `profile` against instance policy (section 8). A
producer never names a repository path or a credential directly.

**Identity.** The producer is derived from the token, never from the body. A
token is scoped to its queues.

**Idempotency.**
- `idempotency_key` is unique per producer and queue for the retention window.
- A resubmission with the same key and the same payload hash returns the
  existing job. The same key with a different payload is rejected with `409`.

**Lifecycle.**

```
queued -> leased -> running -> succeeded | failed | expired | cancelled
running -> draining -> queued              (preempted; see section 7)
queued  -> split_requested -> superseded   (producer replaced it with children)
running -> needs_reconciliation -> reconciled | failed   (task outcome unknown)
succeeded -> unacked_expired               (sensitive payload deleted, section 9)
```

- `parked` is not a state. It is a scheduling flag on a `queued` job
  (section 7), so a parked job still counts toward its producer's outstanding
  limit.
- Terminal states: `succeeded` (once acknowledged), `failed`, `expired`,
  `cancelled`, `superseded`, `reconciled` and `unacked_expired`. Everything
  else counts toward the outstanding limit.
- `split_requested`, `needs_reconciliation` and `unacked_expired` are
  delivered through the same results feed as outputs, as **control results**:
  `{"result_id", "job_id", "control": "<state>", "detail": {...}}`. They are
  acknowledged like any other result.
- **Splitting.**
  1. The producer receives `split_requested` and submits the child jobs.
  2. Each child carries `parent_id` and the deterministic key
     `<parent key>/<index>/<count>`, so a producer that crashes mid-split
     resubmits the same children idempotently.
  3. The coordinator moves the parent to `superseded` once the first child
     arrives, and rejects further children whose `count` differs from the
     first.
  4. A producer that does not support splitting acknowledges the control
     result with `"decline"`. The job stays `queued` and parked (section 7).
- **Reconciliation.** `needs_reconciliation` is resolved by the operator or by
  the producer through `POST /v1/jobs/{id}/reconcile` with `reconciled` or
  `failed`. It is never retried automatically.

- Each lease mints an `attempt_id` and increments the job's fencing
  `generation`.
- Heartbeat, completion and cancellation acknowledgements must carry both. The
  coordinator rejects any that carry a stale generation, so an attempt whose
  lease expired during a partition cannot overwrite a later attempt's result.
- Attempts that fail count against `max_attempts`. Preemptions are counted
  separately and bounded (section 7).

**Results.**
- `GET /v1/jobs/{id}` returns one job.
- `GET /v1/results?queue=&after=<cursor>&limit=` lists results ordered by
  completion sequence, not submission order, with long-poll `wait<=30`.
- A result carries:
  - `result_id`;
  - `output`, validated against `schema` when one was given;
  - `executor`: node, provider and model;
  - `usage`: tokens, wall seconds, and cost in $.
- The producer records `result_id` in the same local transaction as its domain
  writes, and only then calls `POST /v1/jobs/{id}/ack`. A duplicate delivery is
  recognised by `result_id` and dropped.
- Domain validation stays in the producer. Vexa's digit check is an example: a
  rejected answer is the producer's decision, and it resubmits under a new key
  if it wants a retry.

**Bounds.**
- Maximum request and result size, configurable, with a default of 1 MiB. A
  larger request is rejected with `413`.
- Maximum outstanding jobs per producer. When it is reached, submit returns
  `429` with `Retry-After`.
- Listings are paginated.

**Versioning.** `contract` changes only on a breaking change. New optional
fields do not bump it. The repository ships a Python client and a minimal
Rust client crate, and both are tested against shared JSON fixtures.

## 7. Scheduling

**Host state.**
- The node samples:
  - every 2 s while executing;
  - every 30 s at rest.
- Signals:
  - HID idle time from `IOHIDSystem`;
  - power source from `pmset -g ps`;
  - memory pressure from the documented pressure notification.
- A failed read counts as busy.
- The node takes work when all of these hold:
  - idle time is at least the node's threshold (default 5 min);
  - the machine is on AC power;
  - memory pressure is normal and has stayed normal for the recovery window;
  - the job's incremental memory fits the node's budget.
- Idle is measured for the whole machine, whichever user is at the console.
  Guest nodes rely on this.

**Incremental memory.**
- A job for a model that is already resident needs only its KV and execution
  headroom.
- A cold load needs:
  - the weights;
  - KV cache for `num_ctx` times `OLLAMA_NUM_PARALLEL`;
  - buffers;
  - a safety margin.
- A node budget is an enforced admission limit, not a label.

**Release and preemption.**
- These triggers release **idle-only** attempts and stop idle-only admission:
  - user input (idle below 10 s).
- These triggers release **every** attempt, `active_ok` included, and stop all
  admission:
  - switching to battery;
  - pressure at warning or above.
- On release the node:
  1. stops the affected admission;
  2. closes the in-flight request;
  3. marks the attempt `draining`;
  4. waits until the backend is quiet: no active request in the runner;
  5. returns the job to the queue as preempted.
- Closing the HTTP request is only a hint. Ollama cancels cooperatively and
  can defer cancellation during decode, so capacity is released only once the
  backend is quiet.
- **A drain timeout is a drain failure, not a finished drain.**
  - The node stays unavailable.
  - It recovers by unloading the model (`keep_alive: 0`) and, if the runner
    is still busy, by restarting the model server.
  - It confirms the backend is quiet before readmitting anything, and it
    reports the failure in `nodes`.
  - A restart also interrupts the documented direct interactive caller. That
    is accepted and logged.
  - Any late output from the drained attempt is fenced out by its stale
    generation.
- On memory pressure the node also unloads the model before it readmits work.
- Cancel-to-quiet latency is measured on the pinned Ollama build, during both
  prefill and generation, before phase 1 is accepted.

**No starvation by preemption.**
- A job carries a `preemptions` counter.
- After N preemptions (default 3), the producer's batch-shrink hint applies:
  the job is marked `shrink_requested`, and a producer that supports splitting
  resubmits smaller batches.
- If the producer declines, or the job cannot be split, the job is **parked**.
  A parked job is eligible again only when the node's *current* idle period
  has already lasted at least 1.5 times the job's measured p90 runtime, or the
  queue's `parked_min_idle`, whichever is larger. The scheduler hands parked
  work to the idle period in progress, not to history.
- This bounds wasted work: at most N+1 interrupted runs per job. It does not
  bound latency. A parked job on a machine that never stays idle long enough
  waits indefinitely, and `queue` shows it as parked with the idle time it
  needs.
- Wasted runtime per job and per day is recorded and shown in `status`.

**Model residency.**
- Each node keeps one managed model loaded at a time.
- The scheduler prefers jobs for the resident model. It switches only when that
  model's eligible jobs run out, or when another model's oldest job passes its
  queue's maximum age.
- Within a model, jobs are ordered by priority, then by weighted fair share
  across queues, then by age.

**`active_ok` queues.** They are exempt from the idle requirement, and from
release on user input, while the user is active. They still obey the power,
pressure and memory checks, and they run with a concurrency of 1. They do not
preempt an idle-only job that is in flight; they run next.

## 8. Routing, privacy, authorization, quotas

**Executor compatibility comes first.**
- An `inference` job can only go to inference executors that satisfy its
  capability and model requirements, including listed alternatives.
- A `task` can only go to runners its profile authorizes.
- The cost ladder orders compatible executors. It never converts one kind into
  the other.

**Cost ladder, per queue.**
1. A local node where the model is already resident.
2. A local node that can load the model.
3. OpenRouter `:free` endpoints. The limits are 20 requests per minute, and
   either 50 per day or 1000 per day once $10 of credit has been bought. They
   are read from `GET /api/v1/key` (`free_model_daily_requests`).
4. Subscription runners with measured headroom.
5. Paid endpoints, only within the queue's budget. The default budget is $0.

- Each queue sets `max_local_wait` before a job escalates a rung, and can cut
  the ladder short. Translation, for example, is local-only with an infinite
  wait.
- A queue's established lane preferences are expressed in its policy. They are
  not replaced by a global order.

**Privacy.**
- The classes are `public`, `internal`, `personal`, `mail` and `secret`.
- Instance policy maps each class to the executors it may use, and also to the
  node trust classes it may run on: `owner`, `server` or `guest`.
- The coordinator filters candidates by that policy, and the node re-checks it
  before executing.
- Defaults:
  - `public` may go anywhere.
  - `internal` may run on local nodes and subscription runners.
  - `personal`, `mail` and `secret` may use **local executors only**
    (`ollama`, `local-cpu`), and only on `owner` nodes. Both conditions must
    hold.
- A class with no executor permission listed denies everything. External
  processing of a sensitive class needs an explicit override in instance
  policy that names both the class and the executor.
  - Guest nodes receive `public` jobs only unless the policy says otherwise.
- The set of eligible executors must never be empty for a reason that would
  relax privacy. If it is empty, the job waits, or fails with
  `no_eligible_executor`.
- If OpenRouter is ever allowed for a non-public class, requests carry
  `provider.zdr: true`, and only endpoints the policy lists are eligible. ZDR
  filters endpoints; it does not make every free model private.

**Task authorization.**
- Instance policy defines each `profile` as a set of fields:
  - the runner;
  - the credential profile on the node;
  - a sandbox mode;
  - the allowed input roots, as a manifest of files or globs;
  - the privacy ceiling;
  - the nodes the profile may run on.
- Producers are granted profiles per queue.
- A cloud runner receives an isolated workspace holding only the manifest's
  files. It never sees the private instance at large.
- Credentials stay on the node that owns them and are never transported.

**Quotas.**
- Two things are kept apart:
  - the usage the engine measures, in its ledger, per attempt;
  - the headroom the provider reports.
- Provider-reported headroom comes from `/api/v1/key` and from CodexBar
  snapshots. It is keyed by account, model family, window and reset time, and
  carries the sample's freshness and confidence.
- An unknown or stale sample means no headroom, never "assume available".
- Owner reservations are respected: some accounts are reserved for
  interactive use.
- Paid budget is reserved atomically before dispatch, and actual usage is
  settled afterwards, failed and cancelled calls included.
- A quota wall is detected per runner (for example, agy exiting 0 with an empty
  body). It puts that account into cooldown until the reported reset, or a
  default. Jobs are re-planned, not failed.

## 9. Data retention and logging

| Data | Retention |
| --- | --- |
| Inputs and outputs of `mail`, `personal`, `secret` jobs | Stored in a separate database file (`payloads.sqlite3`) with `secure_delete` on, which is never backed up or exported. Deleted at ack; at cancellation; on `failed`, `expired`, `superseded` or `reconciled` after a short inspection window (default 24 h); and for a `succeeded` job that is never acknowledged, after the queue's `unacked_ttl` (default 72 h). That last case moves the job to `unacked_expired` and delivers it as a control result. |
| Other payloads | Stored in the same payload file. Deleted after the queue's retention (default 7 days after ack or terminal state). |
| Errors | Allowlisted fields only: error class, HTTP status, exit code, a truncated provider error code. Never raw provider bodies, generated text or runner output. |
| Backups | Daily `sqlite3 .backup` of the metadata database file only. It holds jobs, attempts, the ledger and quotas, and no content, because content lives in the payload file. On restore, a non-terminal job whose payload is missing becomes `failed` with `payload_lost`, and its producer resubmits it. |
| WAL | Checkpointed on a schedule. Size is monitored and alerts are raised. |
| Logs | Structured, metadata only, under the same allowlist. |

## 10. Failure handling

- **Leases.** A lease has a TTL and is renewed by heartbeat. When it expires,
  the job returns to the queue with a new generation.
- **Tasks.** The outcome of a non-idempotent task (one that wrote files) is
  never replayed automatically after an ambiguous loss. It is marked
  `needs_reconciliation`.
- **Retries.** Backoff is exponential. When attempts run out, the job becomes
  `failed` (dead-letter), with the allowlisted error.
- **Schema violations.** Output that does not match the schema counts as a
  failed attempt, and the job is retried on another compatible executor if one
  exists.
- **Stalled tasks.** A progress guard kills the process group of a `task` whose
  output artifact stops advancing for its profile's stall window. `status`
  shows the age of the last advance, not only liveness.
- **Coordinator down.**
  - Nodes stop taking work.
  - A node that finishes its attempt spools the completion locally, in its own
    SQLite. Content-bearing classes are held only in memory, and lost if the
    node restarts. The completion is retried with its fencing data.
  - Producers treat the coordinator being down as "not yet" and retry the
    submit with the same idempotency key.
- **Durability.** SQLite runs with `synchronous=FULL`, and transactions are
  short. No transaction stays open across a long-poll wait or a network call.

## 11. Security

- The API binds to loopback in phase 1. Later it binds to the host's Tailscale
  address, and requests are checked against enrolled node identities through
  Tailscale's LocalAPI WhoIs as well as the tokens.
- Per-producer and per-node tokens live in the private instance. Nodes are
  enrolled and revoked with a CLI command.
- The repository runs secret scanning in CI and keeps no instance data.
  Examples use placeholder names.

## 12. Deployment lifecycle

- The coordinator and the node are user LaunchAgents. They depend on a
  logged-in user session, and that dependency is documented. Locking the screen
  keeps them running. Logging out stops them.
- An unattended always-on server hosts the coordinator only after it has run
  without a memory panic for an agreed period. On that host the coordinator
  either becomes a LaunchDaemon under a dedicated user, or keeps the
  logged-in-session dependency, stated explicitly.
- Recovery objective: after the coordinator host restarts, no queued or leased
  job is lost. Leases expire and are reissued. Payloads of content-bearing
  classes that were only in flight may have to be resubmitted by their
  producers.

## 13. Observability

The CLI supports `--json` everywhere.

- `status` shows:
  - queue depth and age per queue;
  - useful completed work per hour;
  - wasted preemption time;
  - lease expiries;
  - database busy time and WAL size;
  - quota headroom per account.
- `nodes` shows each node's state and the reason it is not taking work, for
  example `user active 12s`, `on battery`, `pressure warn`, `budget` or
  `draining`. It also shows the resident model and any unexpected runner or
  reload that was observed.
- `queue <name>` shows backlog, throughput, ETA and progress per producer
  batch.
- `costs --since` shows the ledger by provider, queue and day.

## 14. Testing

- Unit and scenario tests in pytest, with a fake clock and simulated host
  state. They cover leases, fencing, preemption with draining and drain failure, parking and splitting,
  wasted-work bounds, residency, routing, privacy filtering and quota reservations.
- Fake Ollama and OpenRouter HTTP servers, including slow cancellation and
  quota-wall behaviour.
- API contract tests against JSON fixtures that the Python and Rust clients
  share.
- Marked manual tests on real hardware:
  - cancel-to-quiet latency;
  - zero reloads across producers, checked with the node's reload counter,
    runner start times and Ollama logs rather than with `/api/ps` alone;
  - useful work completed under repeated interruptions.
- Gates as in the other engines: codeality-py (at most 150 lines per file, one
  unit per file), ruff, mypy, deptry, and coverage with a ratchet.

## 15. Phases

Each phase is independently useful. Details are in the implementation plan.

1. **One machine, one model, one arbiter.**
   - Coordinator and node on the workstation.
   - Ollama executor with pinned options, draining preemption and incremental
     memory admission.
   - Privacy filtering and CLI `status`, `nodes`, `queue`, `jobs`, `cancel`.
   - Python and Rust clients.
   - Every batch qwen caller migrated:
     - Vexa translation;
     - Vexa enrichment, as an `active_ok` queue;
     - Atrium's local drip lane.
   - Vexa's interactive ask stays direct, as the documented exception.
   - Vexa keeps a single result applier and has no direct fallback.
   - Acceptance:
     - 1000 texts translated through the queue;
     - no model reload across the three producers;
     - measured cancel-to-quiet latency;
     - useful work completed under repeated interruptions;
     - `status` showing progress.
2. **Cost ladder.**
   - OpenRouter free executor, ledger and `costs`.
   - Quota snapshots with reservations and wall cooldowns.
   - Atrium's cursor and agy lanes become `task` jobs under authorized
     profiles, and the drip loop is retired.
3. **Absorb the remaining loose scripts.**
   - brain and clips codex/agy passes, with input manifests.
   - Periodic classification loops.
   - Transcription.
4. **More machines.**
   - The coordinator moves to the always-on server once its stability gate
     passes.
   - That server joins as a `server` node, for ONNX work and small models under
     an enforced budget.
   - The guest laptops join as `guest` nodes, once they are reachable, their
     memory is confirmed, their owners agree, and small models have been
     benchmarked.

## 16. Open items carried into planning

- Pin the Ollama version that phase 1 supports, and measure its cancellation
  behaviour.
- Map the memory-pressure source exactly: the documented dispatch source
  rather than a raw sysctl value.
- Specify the host-level test for HID idle under fast user switching and screen
  lock before any guest node is enrolled.
