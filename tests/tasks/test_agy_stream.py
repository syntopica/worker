"""agy stdin streaming: invocation, final-result parsing, usage and wall input."""

import json
from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.agy_answer import agy_answer
from worker.tasks.agy_invocation import agy_invocation
from worker.tasks.agy_no_tools_preface import AGY_NO_TOOLS_PREFACE
from worker.tasks.agy_runner_output import agy_runner_output
from worker.tasks.read_usage import read_usage
from worker.tasks.wall_resets_in_s import wall_resets_in_s

WALL = "Individual quota reached. Please upgrade your subscription. Resets in 1h."


def _stream(*events: dict) -> str:
    return "".join(json.dumps(e) + "\n" for e in events)


def _result(**fields: object) -> dict:
    return {"event": "result", "result": {"status": "SUCCESS", **fields}}


def _profile(**kw: object) -> TaskProfile:
    fields = {
        "name": "p",
        "runner": "agy",
        "model": None,
        "reasoning": None,
        "privacy": frozenset({"internal"}),
        "timeout_s": 600,
        "nodes": None,
        "input_root": None,
    }
    fields.update(kw)
    return TaskProfile(**fields)  # type: ignore[arg-type]


def test_prompt_is_only_on_stdin_as_one_escaped_line() -> None:
    prompt = '--help\n--version \u2028 é "quoted" --print=x'
    invocation = agy_invocation(_profile(model="m"), prompt, Path("/s.json"))
    assert prompt not in " ".join(invocation.argv)
    assert not any(a.startswith("--print=") or a == "-p" for a in invocation.argv)
    formats = ["--input-format", "stream-json", "--output-format", "stream-json"]
    assert invocation.argv[1:5] == formats
    assert "--dangerously-skip-permissions" not in invocation.argv
    assert invocation.argv[-4:] == ["--model", "m", "--json-schema", "/s.json"]
    assert invocation.stdin is not None
    assert invocation.stdin.endswith("\n")
    assert len(invocation.stdin.splitlines()) == 1
    assert invocation.stdin.isascii()
    event = json.loads(invocation.stdin)
    assert event == {
        "event": "user",
        "message": {"content": AGY_NO_TOOLS_PREFACE + prompt},
    }


def test_structured_output_beats_prose_in_final_result() -> None:
    out = _stream(
        {"event": "init"},
        {"event": "step_update", "text": "draft"},
        _result(response="prose", structured_output={"a": 1}),
    )
    assert json.loads(agy_answer(out) or "") == {"a": 1}


def test_final_result_response_and_legacy_envelope() -> None:
    assert agy_answer(_stream(_result(response="hi"))) == "hi"
    assert agy_answer(json.dumps({"response": "old"})) == "old"
    assert json.loads(agy_answer(json.dumps({"structured_output": [1]})) or "") == [1]


def test_truncated_or_malformed_stream_has_no_answer() -> None:
    steps = _stream({"event": "init"}, {"event": "step_update", "text": "partial"})
    assert agy_answer(steps) is None
    assert agy_answer(steps + '{"event":"result","res') is None
    assert agy_answer("") is None
    assert agy_answer("not json") is None


def test_terminal_error_result_is_not_an_answer() -> None:
    failed = _stream({"event": "result", "result": {"status": "ERROR", "response": "x"}})
    assert agy_answer(failed) is None


def test_usage_is_the_final_cumulative_result_only() -> None:
    out = _stream(
        {"event": "step_update", "usage": {"input_tokens": 5, "output_tokens": 5}},
        _result(response="r", usage={"input_tokens": 1, "output_tokens": 2}),
    )
    assert read_usage("agy", out) == {"tokens_in": 1, "tokens_out": 2}
    assert read_usage("agy", _stream({"event": "init"})) == {}
    legacy = json.dumps({"usage": {"input_tokens": 3, "output_tokens": 4}})
    assert read_usage("agy", legacy) == {"tokens_in": 3, "tokens_out": 4}


def test_model_prose_quoting_the_wall_is_not_runner_output() -> None:
    out = _stream(
        {"event": "step_update", "text": WALL},
        _result(response=WALL),
    )
    assert agy_runner_output(out, "", True) == ""
    truncated = _stream({"event": "step_update", "text": WALL})
    assert agy_runner_output(truncated, "", False) == ""


def test_terminal_error_and_stderr_carry_the_wall() -> None:
    failed = _stream({"event": "result", "result": {"status": "ERROR", "error": WALL}})
    text = agy_runner_output(failed, "", False)
    assert wall_resets_in_s("agy", text) == 3600.0
    assert WALL in agy_runner_output("", WALL, False)


def test_legacy_unanswered_envelope_keeps_stdout() -> None:
    legacy = json.dumps({"response": "", "error": WALL})
    assert WALL in agy_runner_output(legacy, "", False)
    assert agy_runner_output(json.dumps({"response": WALL}), "", True) == ""


def test_unicode_line_separator_inside_a_final_answer() -> None:
    out = json.dumps(_result(response="line\u2028answer"), ensure_ascii=False) + "\n"
    assert agy_answer(out) == "line\u2028answer"


def test_whitespace_in_step_events_cannot_fabricate_a_wall() -> None:
    out = '{ "event": "step_update", "text": ' + json.dumps(WALL) + "}\n"
    assert agy_runner_output(out, "", False) == ""
    failed = _stream({"event": "result", "result": {"status": "ERROR", "response": WALL}})
    assert agy_runner_output(failed, "", False) == ""
