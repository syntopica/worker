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

## Continuous TODO, Work Log

Maintain `TODO.md` as the active backlog and `TODO_LOG.md` as the record of
closed work. States: `[ ]` pending, `[~]` partial, `[!]` blocked, `[x]` verified
complete, `[-]` obsolete. Closed items move to `TODO_LOG.md` with date and
evidence.
