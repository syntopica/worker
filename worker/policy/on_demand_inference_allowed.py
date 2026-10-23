"""Keep comparison pins protected unless their substitution was explicitly requested."""

from worker.jobs.candidate import Candidate


def on_demand_inference_allowed(candidate: Candidate, override_pins: bool) -> bool:
    """Original-answer markers never become runnable inference requests."""
    return candidate.kind == "inference" and (
        candidate.pin is None
        or (override_pins and candidate.pin.startswith(("ollama:", "openrouter:", "runner:")))
    )
