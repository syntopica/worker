import json
import time
from pathlib import Path

from tests.tasks.fake_runner import fake_runner
from worker.config.task_profile import TaskProfile
from worker.tasks.codex_failure_text import codex_failure_text
from worker.tasks.read_usage import read_usage
from worker.tasks.run_task import run_task
from worker.tasks.runner_failure_code import runner_failure_code
from worker.tasks.runner_failure_text import runner_failure_text

FIXTURES = Path(__file__).parent / "fixtures"
EVENTS = (FIXTURES / "codex_events.jsonl").read_text()
FAILED = (FIXTURES / "codex_failed_events.jsonl").read_text()


def turn(tokens_in, tokens_out):
    usage = {"input_tokens": tokens_in, "output_tokens": tokens_out}
    return json.dumps({"type": "turn.completed", "usage": usage})


def test_the_captured_turn_maps_input_and_output_tokens():
    assert read_usage("codex", EVENTS) == {"tokens_in": 24525, "tokens_out": 5}


def test_completed_turns_are_summed_and_noise_is_skipped():
    stdout = "\n".join([turn(100, 7), "a warning that is not JSON", "[1]", turn(20, 5), ""])
    assert read_usage("codex", stdout) == {"tokens_in": 120, "tokens_out": 12}
    assert read_usage("codex", turn("many", 3)) == {"tokens_out": 3}


def test_a_failed_run_reports_no_usage():
    assert read_usage("codex", FAILED) == {}
    assert read_usage("codex", "ok\n") == {}


def test_failure_text_keeps_codex_errors_and_drops_warnings_and_answers():
    text = codex_failure_text(FAILED)
    assert text.count("invalid_request_error") == 2
    assert "Under-development" not in text
    answer = json.dumps(
        {"type": "item.completed", "item": {"type": "agent_message", "text": "401"}}
    )
    assert codex_failure_text(answer) == ""


def test_an_auth_failure_in_the_events_is_named():
    stdout = json.dumps({"type": "turn.failed", "error": {"message": "unexpected status 401"}})
    assert runner_failure_code(runner_failure_text("codex", stdout, "")) == "runner_auth"
    assert runner_failure_code(runner_failure_text("cursor", stdout, "")) == "runner_failed"


def profile(command):
    return TaskProfile(
        name="p",
        runner="codex",
        model=None,
        reasoning=None,
        privacy=frozenset({"internal"}),
        timeout_s=30.0,
        nodes=None,
        input_root=None,
        command=command,
    )


def run(body, tmp_path):
    prof = profile(fake_runner(tmp_path, "codex", body))
    the_lease = {"attempt_id": "a", "generation": 1, "input": {"prompt": "p"}}
    return run_task(prof, the_lease, "node-a", lambda: True, time.monotonic, time.sleep)


def test_run_task_passes_json_and_reports_the_turn_usage(tmp_path):
    body = (
        'assert "--json" in args\n'
        'open(args[args.index("-o") + 1], "w").write("ok")\n'
        f"print({EVENTS!r}, end='')"
    )
    report = run(body, tmp_path)
    assert report is not None
    assert report["outcome"] == "succeeded"
    assert report["usage"] == {"tokens_in": 24525, "tokens_out": 5}


def test_run_task_names_a_failure_codex_printed_as_an_event(tmp_path):
    failed = json.dumps({"type": "turn.failed", "error": {"message": "429 Too Many Requests"}})
    report = run(f"print({failed!r})\nsys.exit(1)", tmp_path)
    assert report is not None
    assert (report["outcome"], report["error_code"]) == ("failed", "rate_limited")
