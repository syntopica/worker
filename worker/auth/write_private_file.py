"""Write a file that is mode 0600 from its first byte."""

import os
from pathlib import Path


def write_private_file(path: Path, text: str, *, atomic: bool) -> None:
    """With ``atomic``, write a sibling temp file (O_EXCL) and ``os.replace`` it into place."""
    target = path.with_name(path.name + ".tmp") if atomic else path
    if atomic:
        target.unlink(missing_ok=True)
    flags = os.O_CREAT | os.O_WRONLY | (os.O_EXCL if atomic else os.O_TRUNC)
    with os.fdopen(os.open(target, flags, 0o600), "w") as handle:
        handle.write(text)
    if atomic:
        target.replace(path)
    path.chmod(0o600)
