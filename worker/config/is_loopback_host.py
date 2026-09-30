"""Whether a listen host keeps the API on this machine (spec 12: loopback in phase 1)."""

import ipaddress


def is_loopback_host(host: str) -> bool:
    """True for ``localhost`` and loopback addresses; wildcards and names are refused."""
    if host == "localhost":
        return True
    try:
        return ipaddress.ip_address(host.strip("[]")).is_loopback
    except ValueError:
        return False
