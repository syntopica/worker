"""Executors each privacy class may use unless the instance overrides it (spec 8)."""

DEFAULT_PRIVACY: dict[str, frozenset[str]] = {
    "public": frozenset({"ollama", "local-cpu", "openrouter", "runner"}),
    "internal": frozenset({"ollama", "local-cpu", "runner"}),
    "personal": frozenset({"ollama", "local-cpu"}),
    "mail": frozenset({"ollama", "local-cpu"}),
    "secret": frozenset({"ollama", "local-cpu"}),
}
