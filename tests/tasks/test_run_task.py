import itertools
import time

from tests.tasks.fake_runner import fake_runner
from worker.config.task_profile import TaskProfile
from worker.tasks.run_task import run_task


def profile(runner, command, **extra):
    fields = {
        "name": "p",
        "runner": runner,
        "model": "m-1",
        "reasoning": None,
        "privacy": frozenset({"internal"}),
        "timeout_s": 30.0,
        "nodes": None,
        "input_root": None,
        "command": command,
    }
    fields.update(extra)
    return TaskProfile(**fields)


def lease(prompt="classify", **extra):
    return {"attempt_id": "a", "generation": 1, "input": {"prompt": prompt, **extra}}


def run(prof, the_lease, heartbeat=lambda: True, clock=time.monotonic):
    return run_task(prof, the_lease, "node-a", heartbeat, clock, lambda _s: time.sleep(0.05))


CODEX_OK = """
out = args[args.index("-o") + 1]
seen = sorted(os.listdir("."))
open(out, "w").write(json.dumps({"flags": args[3:7], "seen": seen, "key": os.environ.get("K")}))
"""


def test_codex_runs_read_only_in_a_workspace_holding_only_the_manifest(tmp_path, monkeypatch):
    root = tmp_path / "store"
    (root / "d").mkdir(parents=True)
    (root / "d" / "a.md").write_text("A")
    (root / "b.md").write_text("B")
    monkeypatch.setenv("K", "secret")
    prof = profile(
        "codex",
        fake_runner(tmp_path, "codex", CODEX_OK),
        input_root=root,
        env_unset=frozenset({"K"}),
    )
    report = run(prof, lease(inputs=["d/a.md"], output_schema={"type": "object"}))
    assert report["outcome"] == "succeeded"
    answer = report["output"]["json"]
    assert answer["flags"][1:3] == ["-s", "read-only"]
    assert answer["seen"] == ["d"]
    assert answer["key"] is None
    assert report["executor"] == {"node": "node-a", "provider": "codex", "model": "m-1"}


def test_an_input_escaping_the_root_through_a_link_is_refused(tmp_path):
    root = tmp_path / "store"
    root.mkdir()
    (tmp_path / "outside.md").write_text("x")
    (root / "link.md").symlink_to(tmp_path / "outside.md")
    prof = profile("codex", fake_runner(tmp_path, "codex", ""), input_root=root)
    report = run(prof, lease(inputs=["link.md"]))
    assert (report["outcome"], report["error_code"]) == ("failed", "input_missing")


def test_codex_credit_wall_is_a_quota_wall(tmp_path):
    body = 'print("ERROR: Your workspace is out of credits. Ask your workspace owner to refill in order to continue.")\nsys.exit(1)'
    report = run(profile("codex", fake_runner(tmp_path, "codex", body)), lease())
    assert report["error_code"] == "quota_wall"


def test_agy_reads_structured_output_and_its_silent_wall(tmp_path):
    ok = 'assert "--sandbox" in args and "--dangerously-skip-permissions" not in args\nprint(json.dumps({"response": "prose", "structured_output": {"v": 1}}))'
    report = run(profile("agy", fake_runner(tmp_path, "agy", ok)), lease())
    assert report["output"]["json"] == {"v": 1}
    empty = 'print(json.dumps({"response": None}))'
    report = run(profile("agy", fake_runner(tmp_path, "agy2", empty)), lease())
    assert report["error_code"] == "quota_wall"


def test_cursor_takes_the_prompt_on_stdin_and_the_last_object_of_its_result(tmp_path):
    body = 'result = "narrating {not json} then " + json.dumps({"echo": stdin})\nprint(json.dumps({"is_error": False, "result": result}))'
    report = run(profile("cursor", fake_runner(tmp_path, "cursor", body)), lease("big prompt"))
    assert report["output"]["json"] == {"echo": "big prompt"}


def test_cursor_out_of_usage_is_a_quota_wall_unless_it_answered(tmp_path):
    wall = "ActionRequiredError: Increase limits for faster responses You're out of usage. Switch to Auto, or ask your admin to increase your limit to continue."
    body = f"sys.stderr.write({wall!r})\nsys.exit(1)"
    report = run(profile("cursor", fake_runner(tmp_path, "cursor", body)), lease())
    assert report["error_code"] == "quota_wall"
    quoted = f"print(json.dumps({{'is_error': False, 'result': json.dumps({{'s': {wall!r}}})}}))"
    report = run(profile("cursor", fake_runner(tmp_path, "cursor2", quoted)), lease())
    assert report["outcome"] == "succeeded"


def test_a_runner_that_answers_nothing_fails_without_a_wall(tmp_path):
    report = run(profile("codex", fake_runner(tmp_path, "codex", "sys.exit(3)")), lease())
    assert (report["outcome"], report["error_code"]) == ("failed", "runner_failed")


def test_a_timeout_kills_the_runner(tmp_path):
    prof = profile("codex", fake_runner(tmp_path, "codex", "time.sleep(60)"), timeout_s=0.5)
    started = time.monotonic()
    report = run(prof, lease())
    assert report["error_code"] == "timeout"
    assert time.monotonic() - started < 10


def test_a_fenced_attempt_is_killed_and_not_reported(tmp_path):
    ticks = itertools.count(0, 25)
    prof = profile("codex", fake_runner(tmp_path, "codex", "time.sleep(60)"))
    assert run(prof, lease(), heartbeat=lambda: False, clock=lambda: float(next(ticks))) is None


def test_a_missing_runner_binary_fails_cleanly(tmp_path):
    report = run(profile("codex", str(tmp_path / "absent")), lease())
    assert report["error_code"] == "runner_failed"
