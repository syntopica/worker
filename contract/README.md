# Worker contract v1

Producers talk to the coordinator over HTTP with JSON bodies and a bearer token. The fixtures in `fixtures/` are the canonical examples of each shape; every value in them is a placeholder. The Python client in `worker/client/` is stdlib only and can be vendored.

## Routes

| Method | Path | Caller | Purpose |
| --- | --- | --- | --- |
| POST | `/v1/jobs` | producer | Submit a job. 201 when new, 200 when the idempotency key matched. |
| GET | `/v1/jobs/{id}` | producer | Job state and its latest unacknowledged result. |
| POST | `/v1/jobs/{id}/cancel` | producer | Cancel a job. |
| POST | `/v1/jobs/{id}/ack` | producer | Acknowledge a result, or decline a `split_requested`. |
| GET | `/v1/results` | producer | Unacknowledged results of a queue (`queue`, `after`, `limit`, `wait`). |
| POST | `/v1/leases` | node | Ask for work. 200 with a lease, 204 when nothing is eligible. |
| POST | `/v1/attempts/{id}/heartbeat` | node | Extend a lease. |
| POST | `/v1/attempts/{id}/complete` | node | Report the outcome of an attempt. |
| POST | `/v1/nodes/{name}/report` | node | Report host state. |
| GET | `/v1/status` | any | Queue and node status. |

Errors are `{"error": "<code>"}` with the HTTP status. `error_stale_attempt.json` is an example.

## Lifecycle

1. Submit a job with a stable `idempotency_key`. Resubmitting the same body returns the same job; a different body under the same key is `idempotency_conflict`.
2. A node leases the job, runs it and completes the attempt. A lease that stops heartbeating expires and the job is retried up to `max_attempts`.
3. The outcome appears as a result row in `GET /v1/results`, ordered by `seq`.
4. Record `result_id` transactionally with your own domain writes, then ack. Acknowledging the same result twice is harmless, so a crash between the two steps only repeats the ack.
5. An unacknowledged result expires after its retention window and is reported as `unacked_expired`.

## Control results

A result row with a non-null `control` carries no output, executor or usage, and its `detail` depends on the control.

- `split_requested`: the job was preempted repeatedly. `detail` is `{"preemptions": <count>}`. Submit split children, or decline through `ack` with `decline: true` to park the job.
- `failed`: attempts were exhausted, the lease was lost or preemption was exhausted. `detail` is `{"error": "<code>"}`. See `result_failed.json`.
- `expired`: the deadline passed before completion. `detail` is `{"error": "deadline"}`.
- `unacked_expired`: a result was never acknowledged in time. `detail` is `{}`.

The codes a `failed` result can carry come from a fixed vocabulary: `lease_lost`, `preemption_exhausted`, `schema_violation`, `executor_error`, `transport_error`, `bad_response` and `unknown_model`, plus `http_<status>` for a non-200 answer from the backend, where `<status>` is the numeric status only. A code never contains provider or generated text.

## Limits and scope

- Phase 1a accepts `kind: "inference"` only. `kind: "task"` is rejected with `400 unsupported_kind`.
- `limit` on `GET /v1/results` is capped at 100.
- `wait` on `GET /v1/results` is capped at 30 seconds.
- The `needs_reconciliation` state and its `/reconcile` route are deferred to phase 2.

## Split children

A child of a split uses the idempotency key `<parent key>/<index>/<count>` and carries `parent_id`.
