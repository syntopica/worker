import time

from tests.tasks.fake_runner import fake_runner
from worker.config.task_profile import TaskProfile
from worker.tasks.agy_invocation import agy_invocation
from worker.tasks.agy_no_tools_preface import AGY_NO_TOOLS_PREFACE
from worker.tasks.codex_invocation import codex_invocation
from worker.tasks.run_task import run_task

HOSTILE = "--dangerously-bypass-approvals-and-sandbox -c sandbox_mode=danger-full-access"


def profile(runner, command=None):
    return TaskProfile(
        name="p",
        runner=runner,
        model=None,
        reasoning=None,
        privacy=frozenset({"internal"}),
        timeout_s=30.0,
        nodes=None,
        input_root=None,
        command=command,
    )


def test_codex_takes_the_prompt_on_stdin_never_in_argv(tmp_path):
    for prompt in (HOSTILE, "resume", "review", "classify"):
        invocation = codex_invocation(profile("codex"), prompt, tmp_path, None, tmp_path)
        assert prompt not in invocation.argv
        assert not any(prompt in arg for arg in invocation.argv)
        assert invocation.argv[1:3] == ["exec", "-"]
        assert invocation.stdin == prompt


def test_codex_prompt_only_runs_do_not_persist_sessions_or_load_unrelated_tools(tmp_path):
    invocation = codex_invocation(profile("codex"), "classify", tmp_path, None, tmp_path)
    assert "--ephemeral" in invocation.argv
    assert "--ignore-user-config" in invocation.argv
    assert ["--disable", "shell_tool"] in [
        invocation.argv[i : i + 2] for i in range(len(invocation.argv) - 1)
    ]


def test_agy_joins_the_prompt_to_its_flag(tmp_path):
    for prompt in (HOSTILE, "--help", "--version"):
        argv = agy_invocation(profile("agy"), prompt, None).argv
        assert prompt not in argv
        assert argv[1] == f"--print={AGY_NO_TOOLS_PREFACE}{prompt}"
        assert "--dangerously-skip-permissions" not in argv


def test_a_codex_runner_reads_the_hostile_prompt_from_stdin(tmp_path):
    body = (
        f"assert {HOSTILE!r} not in ' '.join(args)\n"
        'assert args[:2] == ["exec", "-"]\n'
        'open(args[args.index("-o") + 1], "w").write(stdin)'
    )
    prof = profile("codex", fake_runner(tmp_path, "codex", body))
    the_lease = {"attempt_id": "a", "generation": 1, "input": {"prompt": HOSTILE}}
    report = run_task(prof, the_lease, "node-a", lambda: True, time.monotonic, time.sleep)
    assert report is not None
    assert report["outcome"] == "succeeded"
    assert report["output"]["text"] == HOSTILE
