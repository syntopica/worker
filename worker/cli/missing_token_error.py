"""A token file that does not exist."""


class MissingTokenError(Exception):
    """Raised when state/tokens/<name>.token is absent."""
