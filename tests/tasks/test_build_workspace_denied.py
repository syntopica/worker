import pytest

from worker.config.parse_task_profile import parse_task_profile
from worker.tasks.build_workspace import build_workspace
from worker.tasks.workspace_error import WorkspaceError


def test_the_worker_state_is_never_copied_even_under_a_broad_root(tmp_path):
    instance = tmp_path / "instance"
    (instance / "worker" / "state" / "tokens").mkdir(parents=True)
    (instance / "worker" / "state" / "tokens" / "a.token").write_text("t")
    (instance / "notes.md").write_text("n")
    profile = parse_task_profile("p", {"runner": "codex", "input_root": ".."}, instance / "worker")
    workspace = tmp_path / "ws"
    workspace.mkdir()
    build_workspace(profile.input_root, ["notes.md"], workspace, profile.denied_root)
    assert (workspace / "notes.md").read_text() == "n"
    with pytest.raises(WorkspaceError, match="input_denied"):
        build_workspace(
            profile.input_root, ["worker/state/tokens/a.token"], workspace, profile.denied_root
        )
