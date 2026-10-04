import subprocess
import sys

import pytest

from worker.node.run_command import run_command


def test_success_returns_stdout(capsys: pytest.CaptureFixture[str]) -> None:
    assert run_command([sys.executable, "-c", "print('ok')"]) == "ok\n"
    assert capsys.readouterr().err == ""


def test_non_zero_exit_logs_status_only(capsys: pytest.CaptureFixture[str]) -> None:
    args = [sys.executable, "-c", "import sys; print('secret'); sys.exit(3)"]
    assert run_command(args) is None
    err = capsys.readouterr().err
    assert "exit 3" in err
    assert "secret" not in err


def test_timeout_is_logged_as_timeout(
    monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]
) -> None:
    def slow(*_args: object, **_kwargs: object) -> None:
        raise subprocess.TimeoutExpired(cmd="pmset", timeout=5.0)

    monkeypatch.setattr(subprocess, "run", slow)
    assert run_command(["/usr/bin/pmset", "-g", "ps"]) is None
    assert capsys.readouterr().err.strip() == "worker: command failed: pmset timeout 5s"


def test_missing_binary_logs_os_error_name(capsys: pytest.CaptureFixture[str]) -> None:
    assert run_command(["/nonexistent/ioreg"]) is None
    assert capsys.readouterr().err.strip() == "worker: command failed: ioreg oserror ENOENT"
