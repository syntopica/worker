"""Whether a request path fits a route pattern."""


def path_matches(pattern: list[str], parts: list[str]) -> bool:
    """Same segment count; each ``*`` stands for exactly one segment, the rest are literal."""
    return len(pattern) == len(parts) and all(
        want in ("*", got) for want, got in zip(pattern, parts, strict=True)
    )
