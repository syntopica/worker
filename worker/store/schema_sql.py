"""Metadata tables (main) and content tables (p, the payload file) - spec 9."""

SCHEMA = """
CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  producer TEXT NOT NULL,
  queue TEXT NOT NULL,
  kind TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  payload_hash TEXT NOT NULL,
  priority INTEGER NOT NULL,
  privacy TEXT NOT NULL,
  model TEXT NOT NULL,
  state TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL,
  preemptions INTEGER NOT NULL DEFAULT 0,
  parked INTEGER NOT NULL DEFAULT 0,
  parked_runs INTEGER NOT NULL DEFAULT 0,
  parked_min_idle_s REAL NOT NULL DEFAULT 0,
  generation INTEGER NOT NULL DEFAULT 0,
  lease_node TEXT,
  lease_attempt TEXT,
  lease_expires REAL,
  not_before REAL NOT NULL,
  deadline REAL,
  parent_id TEXT,
  split_count INTEGER,
  error TEXT,
  created REAL NOT NULL,
  updated REAL NOT NULL,
  finished REAL,
  acked REAL,
  payloads_deleted INTEGER NOT NULL DEFAULT 0,
  tier TEXT NOT NULL DEFAULT 'basic',
  UNIQUE (producer, queue, idempotency_key)
);
CREATE INDEX IF NOT EXISTS jobs_ready ON jobs (state, not_before);
CREATE INDEX IF NOT EXISTS jobs_attempt ON jobs (lease_attempt);
CREATE TABLE IF NOT EXISTS attempts (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  generation INTEGER NOT NULL,
  node TEXT NOT NULL,
  started REAL NOT NULL,
  ended REAL,
  outcome TEXT,
  error TEXT,
  wall_s REAL,
  tokens_in INTEGER,
  tokens_out INTEGER,
  provider TEXT,
  cost_usd REAL,
  model TEXT
);
CREATE INDEX IF NOT EXISTS attempts_started ON attempts (started);
CREATE TABLE IF NOT EXISTS results (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  result_id TEXT NOT NULL UNIQUE,
  job_id TEXT NOT NULL,
  producer TEXT NOT NULL,
  queue TEXT NOT NULL,
  control TEXT,
  detail TEXT,
  executor TEXT,
  usage TEXT,
  created REAL NOT NULL,
  acked REAL,
  rating TEXT
);
CREATE TABLE IF NOT EXISTS nodes (name TEXT PRIMARY KEY, report TEXT NOT NULL, updated REAL NOT NULL);
CREATE TABLE IF NOT EXISTS cooldowns (runner TEXT PRIMARY KEY, until REAL NOT NULL);
CREATE TABLE IF NOT EXISTS p.inputs (job_id TEXT PRIMARY KEY, body TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS p.outputs (result_id TEXT PRIMARY KEY, body TEXT NOT NULL);
"""
