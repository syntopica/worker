"""Read a saved bearer token."""

from pathlib import Path

from worker.cli.missing_token_error import MissingTokenError


def read_token(state: Path, name: str) -> str:
    """Return the token, or raise MissingTokenError naming the file's principal only."""
    path = state / "tokens" / f"{name}.token"
    try:
        return path.read_text().strip()
    except FileNotFoundError:
        raise MissingTokenError(
            f"no token for '{name}' (run: worker token add --name {name})"
        ) from None
