"""Restart the model server's LaunchAgent as the last drain recovery step."""

import os

from worker.node.run_command import run_command


def restart_ollama(label: str) -> bool:
    """``launchctl kickstart -k`` on the user's GUI domain."""
    return (
        run_command(["/bin/launchctl", "kickstart", "-k", f"gui/{os.getuid()}/{label}"], timeout=30)
        is not None
    )
