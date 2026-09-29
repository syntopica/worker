"""Whether the machine is on AC power, from ``pmset -g ps``."""


def parse_on_ac(text: str) -> bool | None:
    """None when the output names neither source."""
    first = text.splitlines()[0] if text else ""
    if "AC Power" in first:
        return True
    if "Battery Power" in first:
        return False
    return None
