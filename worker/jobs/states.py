"""Job state and privacy vocabularies (spec 6 and 8)."""

LIVE = ("leased", "running", "draining")
PRIVACY_CLASSES = ("public", "internal", "personal", "mail", "secret")
SENSITIVE = ("personal", "mail", "secret")
NOT_OUTSTANDING = ("superseded", "cancelled", "unacked_expired")
CONTROL_TERMINAL = ("failed", "expired", "unacked_expired")
RELEASABLE = ("failed", "expired", "cancelled", "superseded", "unacked_expired", "reconciled")
TIERS = ("basic", "strong")
RATINGS = ("good", "edited", "discarded")
JOB_STATES = (
    "queued",
    "leased",
    "running",
    "draining",
    "split_requested",
    "superseded",
    "needs_reconciliation",
    "reconciled",
    "succeeded",
    "failed",
    "expired",
    "cancelled",
    "unacked_expired",
)
RETRYABLE = ("failed", "expired", "cancelled")
SAMPLING_PRODUCERS = ("_shadow", "_judge")
