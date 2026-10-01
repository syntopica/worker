"""Choose the host sampler for the platform a node runs on."""

import functools
import sys
from collections.abc import Callable

from worker.config.node_policy import NodePolicy
from worker.node.bridge_unreadable import BridgeUnreadable
from worker.node.hold_power_source import HoldPowerSource
from worker.node.host_state import HostState
from worker.node.linux_host_sampler import LinuxHostSampler
from worker.node.sample_host_state import sample_host_state


def default_host_sampler(node: NodePolicy, platform: str = sys.platform) -> Callable[[], HostState]:
    """Linux reads load and memory from /proc; anything else is the macOS reader."""
    if platform.startswith("linux"):
        return LinuxHostSampler(node.max_load, node.min_free_pct)
    return HoldPowerSource(
        BridgeUnreadable(functools.partial(sample_host_state, min_free_pct=node.min_free_pct))
    )
