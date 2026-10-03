"""Read a runner's ``model_windows``: window label to the model fragments it meters."""

from typing import Any


def parse_model_windows(runner: str, raw: Any) -> dict[str, tuple[str, ...]]:
    """Raise ValueError unless every window names a non-empty list of strings."""
    if not isinstance(raw, dict) or not all(
        isinstance(fragments, list) and fragments and all(isinstance(f, str) for f in fragments)
        for fragments in raw.values()
    ):
        raise ValueError(f"runner {runner}: model_windows must map windows to model fragments")
    return {str(window): tuple(fragments) for window, fragments in raw.items()}
