"""Write an executable stand-in for a runner CLI."""

import stat
import sys
from pathlib import Path


def fake_runner(tmp_path: Path, name: str, body: str) -> str:
    """A Python script named ``name``; ``body`` sees ``args``, ``stdin`` and ``os``."""
    script = tmp_path / name
    script.write_text(
        f"#!{sys.executable}\nimport json, os, sys, time\n"
        "args = sys.argv[1:]\nstdin = sys.stdin.read()\n" + body
    )
    script.chmod(script.stat().st_mode | stat.S_IEXEC)
    return str(script)
