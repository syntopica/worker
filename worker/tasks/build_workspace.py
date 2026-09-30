"""Copy a task's manifest into a fresh workspace (spec 8: nothing else is visible)."""

import shutil
from pathlib import Path

from worker.tasks.max_workspace_bytes import MAX_WORKSPACE_BYTES
from worker.tasks.workspace_error import WorkspaceError


def build_workspace(
    root: Path | None, inputs: list[str], workspace: Path, denied: Path | None = None
) -> None:
    """Copy each entry to the same relative path; raise WorkspaceError on any doubt.

    An entry must resolve inside ``root`` after symlinks, so a link planted in
    the input store cannot hand the runner a file from elsewhere, and outside
    ``denied`` (the worker's state), which a broad ``root`` may contain.
    """
    if inputs and root is None:
        raise WorkspaceError("inputs_not_allowed")
    total = 0
    for entry in inputs:
        source = (root / entry).resolve() if root is not None else Path()
        if root is None or not source.is_relative_to(root) or not source.is_file():
            raise WorkspaceError("input_missing")
        if denied is not None and source.is_relative_to(denied):
            raise WorkspaceError("input_denied")
        total += source.stat().st_size
        if total > MAX_WORKSPACE_BYTES:
            raise WorkspaceError("input_too_large")
        target = workspace / entry
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
