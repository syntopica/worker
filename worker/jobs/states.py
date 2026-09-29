"""Job state and privacy vocabularies (spec 6 and 8)."""

LIVE = ("leased", "running", "draining")
PRIVACY_CLASSES = ("public", "internal", "personal", "mail", "secret")
SENSITIVE = ("personal", "mail", "secret")
NOT_OUTSTANDING = ("superseded", "cancelled", "unacked_expired")
CONTROL_TERMINAL = ("failed", "expired", "unacked_expired")
