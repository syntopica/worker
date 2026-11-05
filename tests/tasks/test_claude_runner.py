"""The claude runner: CLI flags, envelope answers, walls and its account environment."""

import json
import time
from pathlib import Path

import pytest

from tests.tasks.fake_runner import fake_runner
from worker.config.parse_task_profile import parse_task_profile
from worker.config.task_profile import TaskProfile
from worker.tasks.claude_answer import claude_answer
from worker.tasks.claude_invocation import claude_invocation
from worker.tasks.is_quota_wall import is_quota_wall
from worker.tasks.read_usage import read_usage
from worker.tasks.run_task import run_task

ENVELOPE = {
    "type": "result",
    "is_error": False,
    "result": '{"label": "ok"}',
    "structured_output": {"label": "ok"},
    "usage": {"input_tokens": 12, "output_tokens": 5},
}


def profile(command: str | None = None, env: tuple[tuple[str, str], ...] = ()) -> TaskProfile:
    return TaskProfile(
        "c", "claude", "model-x", None, frozenset({"internal"}), 30, None, None, command, env=env
    )


def test_the_invocation_has_no_tools_settings_or_session(tmp_path: Path) -> None:
    schema = tmp_path / "s.json"
    schema.write_text('{"type": "object"}')
    got = claude_invocation(profile(), "the prompt", schema)
    assert got.argv[:4] == ["claude", "-p", "--model", "model-x"]
    assert got.argv[got.argv.index("--tools") + 1] == ""
    assert got.argv[got.argv.index("--setting-sources") + 1] == ""
    assert "--no-session-persistence" in got.argv
    assert "--strict-mcp-config" in got.argv
    assert got.argv[got.argv.index("--json-schema") + 1] == '{"type": "object"}'
    assert got.stdin == "the prompt"
    assert "the prompt" not in got.argv
    assert "--json-schema" not in claude_invocation(profile(), "p", None).argv


def test_the_envelope_yields_structured_output_then_result_and_usage() -> None:
    assert claude_answer(json.dumps(ENVELOPE)) == '{"label": "ok"}'
    assert claude_answer(json.dumps({"result": "plain"})) == "plain"
    assert claude_answer(json.dumps({**ENVELOPE, "is_error": True})) is None
    assert claude_answer("not json") is None
    assert read_usage("claude", json.dumps(ENVELOPE)) == {"tokens_in": 12, "tokens_out": 5}


@pytest.mark.parametrize(
    "text",
    [
        '{"is_error": true, "result": "Claude AI usage limit reached|1793900000"}',
        '{"is_error": true, "result": "You\'ve hit your limit \\u00b7 resets 9pm"}',
        '{"is_error": true, "result": "5-hour limit reached \\u2219 resets 4am"}',
        '{"is_error": true, "api_error_status": 429, "result": "API Error"}',
    ],
)
def test_a_limit_without_an_answer_is_a_wall(text: str) -> None:
    assert is_quota_wall("claude", 1, text, None)


def test_an_answer_quoting_a_limit_is_never_a_wall() -> None:
    assert not is_quota_wall("claude", 0, "You've hit your limit", "a quoted limit")
    assert not is_quota_wall("claude", 1, '{"is_error": true, "result": "boom"}', None)


def test_a_claude_profile_names_its_model_and_expands_its_env(tmp_path: Path) -> None:
    got = parse_task_profile(
        "c", {"runner": "claude", "model": "m", "env": {"CLAUDE_CONFIG_DIR": "~/.x"}}, tmp_path
    )
    assert got.env == (("CLAUDE_CONFIG_DIR", str(Path.home() / ".x")),)
    with pytest.raises(ValueError, match="name a model"):
        parse_task_profile("c", {"runner": "claude"}, tmp_path)
    with pytest.raises(ValueError, match="env must map"):
        parse_task_profile("c", {"runner": "claude", "model": "m", "env": {"A": 1}}, tmp_path)


def test_a_run_answers_under_the_profile_environment(tmp_path: Path) -> None:
    body = (
        'assert "--tools" in args\n'
        "print(json.dumps({'is_error': False, 'structured_output': "
        "{'dir': os.environ.get('CLAUDE_CONFIG_DIR'), 'stdin': stdin}, "
        "'usage': {'input_tokens': 3, 'output_tokens': 4}}))\n"
    )
    command = fake_runner(tmp_path, "claude", body)
    prof = profile(command, env=(("CLAUDE_CONFIG_DIR", "/cfg"),))
    lease = {"attempt_id": "a", "generation": 1, "input": {"prompt": "hello"}}
    report = run_task(prof, lease, "node-a", lambda: True, time.monotonic, lambda _s: None)
    assert report is not None
    assert report["outcome"] == "succeeded"
    assert json.loads(report["output"]["text"]) == {"dir": "/cfg", "stdin": "hello"}
