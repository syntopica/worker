"""A manifest that cannot be copied into a task workspace."""


class WorkspaceError(Exception):
    """Carries no path: the message is an allowlisted code only."""
