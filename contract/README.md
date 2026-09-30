# Worker contract v1

Producers talk to the coordinator over HTTP with JSON bodies and a bearer token. The fixtures in `fixtures/` are the canonical examples of each shape; every value in them is a placeholder. The Python client in `worker/client/` is stdlib only and can be vendored.

## Routes

| Method | Path | Caller | Purpose |
| --- | --- | --- | --- |
| POST | `/v1/jobs` | producer | Submit a job. 201 when new, 200 when the idempotency key matched. |
| GET | `/v1/jobs/{id}` | producer | Job state and its latest unacknowledged result. |
| POST | `/v1/jobs/{id}/cancel` | producer | Cancel a job. |
| POST | `/v1/jobs/{id}/ack` | producer | Acknowledge a result, or decline a `split_requested`. May carry `rating`. |
| GET | `/v1/results` | producer | Unacknowledged results of a queue (`queue`, `after`, `limit`, `wait`). |
| POST | `/v1/leases` | node | Ask for work. 200 with a lease, 204 when nothing is eligible. |
| POST | `/v1/attempts/{id}/heartbeat` | node | Extend a lease. |
| POST | `/v1/attempts/{id}/complete` | node | Report the outcome of an attempt. |
| POST | `/v1/nodes/{name}/report` | node | Report host state. |
| GET | `/v1/status` | admin | Queue and node status. |
| GET | `/v1/costs` | admin | Attempts, tokens, cost and wall time by day, provider and queue (`days`). |
| GET | `/v1/quality` | admin | Outcomes and producer ratings by queue, tier, provider and model (`days`). |

Errors are `{"error": "<code>"}` with the HTTP status. `error_stale_attempt.json` is an example.

## Quality tier and rating

A submit may carry `tier`: `basic` (the default) or `strong`; anything else is `unknown_tier`. The queue's configuration decides what a strong job runs on (models tried before your own list, or a stronger task profile); a queue that maps no tier runs it as `basic`. The tier never changes the privacy rules. When you ack a result you may add `rating`: `good`, `edited` (used after correction) or `discarded` (unusable); anything else is `unknown_rating`. A later ack of the same result may change it. Ratings are how the owner learns whether a cheap model is good enough for your queue, so rate when you can.

## Lifecycle

1. Submit a job with a stable `idempotency_key`. Resubmitting the same body returns the same job; a different body under the same key is `idempotency_conflict`.
2. A node leases the job, runs it and completes the attempt. A lease that stops heartbeating expires and the job is retried up to `max_attempts`.
3. The outcome appears as a result row in `GET /v1/results`, ordered by `seq`.
4. Record `result_id` transactionally with your own domain writes, then ack. Acknowledging the same result twice is harmless, so a crash between the two steps only repeats the ack.
5. An unacknowledged result expires after its retention window and is reported as `unacked_expired`.
6. A finished job is forgotten once its retention has passed after its ack or last transition; its idempotency key can then be reused for a new job.

## Control results

A result row with a non-null `control` carries no output, executor or usage, and its `detail` depends on the control.

- `split_requested`: the job was preempted repeatedly. `detail` is `{"preemptions": <count>}`. Submit split children, or decline through `ack` with `decline: true` to park the job.
- `failed`: attempts were exhausted, the lease was lost or preemption was exhausted. `detail` is `{"error": "<code>"}`. See `result_failed.json`.
- `expired`: the deadline passed before completion. `detail` is `{"error": "deadline"}`.
- `unacked_expired`: a result was never acknowledged in time. `detail` is `{}`.

The codes a `failed` result can carry come from a fixed vocabulary: `lease_lost`, `preemption_exhausted`, `schema_violation`, `executor_error`, `transport_error`, `bad_response`, `unknown_model`, `payload_lost` (the stored input is gone; resubmit the job) and `node_error` (the node failed unexpectedly), plus `http_<status>` for a non-200 answer from the backend, where `<status>` is the numeric status only. Tasks add `timeout`, `runner_failed` (the runner exited non-zero with no answer, or could not start), `no_output` (it exited 0 with no answer), `input_missing`, `input_denied` (the entry is inside the worker's own state), `input_too_large`, `inputs_not_allowed` and `unknown_profile`. Any kind may end `privacy_refused` when the node's own policy forbids running the job there. A `schema_violation` result's `detail` may add `schema_path`, a JSON pointer into the producer's schema naming the first rule the output broke. A `quota_wall` never reaches a result: the job is requeued after the runner's cooldown. Nor does `rate_limited`, a free remote endpoint's 429, or `node_shutdown`, the node stopping mid-attempt: the job is requeued at once without charging an attempt. A `split_requested` left unanswered longer than the queue's unacknowledged TTL is declined for you: the job is parked. While a queued task's runner is resting, `GET /v1/jobs/{id}` reports the cooldown's end as `cooling_until` (epoch seconds, otherwise null), so a producer can stop waiting and collect the job later by its key. A code never contains provider or generated text.

## Input

An inference `input` carries `messages`, a non-empty list of `{"role": <string>, "content": <string>}`, and optionally `options` (an object), `schema` (an object) and `format` (only `"json"`). Anything else is rejected with `400 bad_input`.

A task `input` carries `profile` and `prompt` (non-empty strings), optionally `runner` (when given it must be the profile's), `inputs` (at most 64 paths relative to the profile's input root, without `..`) and `output_schema` (an object, validated like `schema`). Anything else is `400 bad_input`. The queue must grant the profile (`403 profile_not_granted`), `runner` must be the profile's (`400 runner_mismatch`), the privacy class must be one the profile accepts (`403 privacy_not_allowed`), and inputs need a profile with an input root (`400 inputs_not_allowed`). `submit_task.json` is an example. A task's result has the same `output` shape as inference, `{"text", "json"}`, and its `executor.provider` is the runner.

## Limits and scope

- `kind: "task"` runs read-only tasks only: the product is the final message, never a file.
- `limit` on `GET /v1/results` is capped at 100.
- `wait` on `GET /v1/results` is capped at 30 seconds.
- The `needs_reconciliation` state and its `/reconcile` route are deferred to phase 2.

## Split children

A child of a split uses the idempotency key `<parent key>/<index>/<count>` and carries `parent_id`.
