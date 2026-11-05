"""Whether an OpenRouter model id may be called without a budget."""

from worker.config.free_router_model import FREE_ROUTER_MODEL


def is_free_openrouter_model(model: object) -> bool:
    """A ``:free`` endpoint, or the free router that picks one."""
    return isinstance(model, str) and (model.endswith(":free") or model == FREE_ROUTER_MODEL)
