"""Node trust classes each privacy class may run on unless overridden (spec 8)."""

DEFAULT_TRUST: dict[str, frozenset[str]] = {
    "public": frozenset({"owner", "server", "guest"}),
    "internal": frozenset({"owner", "server"}),
    "personal": frozenset({"owner"}),
    "mail": frozenset({"owner"}),
    "secret": frozenset({"owner"}),
}
