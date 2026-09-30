"""The route a queue maps a quality tier to, if any."""

from worker.config.queue_policy import QueuePolicy
from worker.config.tier_route import TierRoute


def tier_route_for(queue: QueuePolicy, tier: str) -> TierRoute:
    """An unmapped tier behaves as ``basic`` with nothing mapped: an empty route."""
    return dict(queue.tiers).get(tier, TierRoute())
