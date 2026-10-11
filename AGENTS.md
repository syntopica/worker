# worker - repository instructions

## What this is

A background AI job processor: a coordinator (SQLite, HTTP API) and one node per
machine that executes jobs only while its machine is idle. The design of record
is `docs/superpowers/specs/2026-09-29-worker-design.md`; change it before
changing behaviour it specifies.

## Hard boundaries

- **Public repository, engine only.** Never commit instance data: machine or
  host names, addresses, people's names, tokens, provider keys, queue contents,
  logs or database files. Examples use placeholders. Instance configuration is
  resolved from `SYNTOPICA_DATA` / `syntopica.config.json`.
- **Content-bearing payloads never reach logs, errors or backups.** Errors carry
  allowlisted fields only (spec section 9).
- **Privacy never relaxes to find an executor.** An empty eligible set waits or
  fails `no_eligible_executor`.
- **Every completion is fenced** by attempt id and generation; never accept a
  stale one.

## Conventions

- Python with uv, snake_case, ruff-formatted, double quotes, mypy, pytest.
  Tests are `tests/test_*.py`.
- One exported unit and one responsibility per file (codeality-py: at most 150
  lines per file); every dependency an explicit import.
- All code, comments, docs and commit messages in English.
- Only `main`; commit and push when the gate is green.

<!-- tadu:rule v1 -->
## Continuous TODO, Work Log, and History Coverage

Maintain `TODO.md` as the active backlog and `TODO_LOG.md` as the searchable
record of closed work. Use `TODO_HISTORY_INDEX.jsonl` to avoid parsing unchanged
conversations more than once.

- Read `TODO.md` at the beginning and end of related work. Search `TODO_LOG.md`
  before reopening an old task or repeating a previous solution.
- Record actionable bugs, risks, blockers, deferred work, missing tests,
  validation, documentation, and product improvements as they are discovered.
- Update an existing entry instead of creating a duplicate. Keep entries concise
  and under the most relevant category.
- Use `[ ]` pending, `[~]` partial or unverified, `[!]` blocked, `[x]` verified
  complete, `[-]` obsolete or superseded. Keep blockers in `TODO.md` and name
  the smallest action required to unblock them.
- When work becomes `[x]` or `[-]`, append a dated entry with concise result and
  evidence to `TODO_LOG.md`, then remove it from the active backlog. Keep one
  log file, grouped by year and month.
- Before reviewing past conversations, consult the history index and skip
  unchanged records already marked `complete` or `irrelevant`. Update a record
  only after its findings are reconciled; interrupted work stays `partial`.
- Do not interrupt the active task for unrelated non-critical work, and do not
  implement unrelated TODO items unless requested. Immediately report critical
  security, destructive, or data-loss findings.

Format: tadu spec v1. Workflows: the tadu-create and tadu-work skills.
<!-- /tadu:rule -->

## TODO conventions specific to this repository

Both files, the spec and commit messages are public: write engine behaviour
and engine commits only. Instance operations - machines, queue names and
contents, live results, quota readings, other repositories' commits, owner
quotes - go to the private instance's `worker/TODO.md` and
`worker/TODO_LOG.md` under `SYNTOPICA_DATA`.
