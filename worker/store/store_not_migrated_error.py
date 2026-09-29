"""Raised when a store is opened before it was migrated to the current schema."""


class StoreNotMigratedError(RuntimeError):
    """``worker serve`` migrates the store at startup; nothing else may run before it."""

    def __init__(self, found: int, wanted: int) -> None:
        super().__init__(f"store schema version {found} is older than {wanted}; run worker serve")
        self.found = found
        self.wanted = wanted
