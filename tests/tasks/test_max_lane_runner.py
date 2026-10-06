import json
from pathlib import Path

from worker.config.task_profile import TaskProfile
from worker.tasks.drain_on_demand import drain_on_demand
from worker.tasks.is_quota_wall import is_quota_wall
from worker.tasks.max_lane_answer import max_lane_answer
from worker.tasks.max_lane_invocation import max_lane_invocation
from worker.tasks.read_usage import read_usage

ENVELOPE = (Path(__file__).parent / "fixtures" / "max_lane_envelope.json").read_text()
WALL = (
    "max-lane-run: Max lane request failed for Keychain service"
    " Claude Code-credentials-36d7ba4b with HTTP status 429.\n"
)


def profile():
    return TaskProfile("max", "max-lane", "claude-x", None, frozenset(), 60, None, None)


def test_the_invocation_passes_the_prompt_on_stdin_and_the_schema_by_path():
    got = max_lane_invocation(profile(), "the prompt", Path("/s.json"))
    assert got.argv == ["max-lane-run", "--model", "claude-x", "--schema", "/s.json"]
    assert got.stdin == "the prompt"
    assert got.answer_file is None


def test_the_envelope_yields_its_answer_and_usage():
    assert max_lane_answer(ENVELOPE) == "ok"
    assert max_lane_answer(json.dumps({"answer": {"a": 1}})) == '{"a": 1}'
    assert max_lane_answer("not json") is None
    assert max_lane_answer(json.dumps({"answer": 3})) is None
    assert read_usage("max-lane", ENVELOPE) == {"tokens_in": 626, "tokens_out": 34}


def test_every_account_at_429_is_a_wall_but_an_answer_never_is():
    assert is_quota_wall("max-lane", 1, WALL, None)
    assert not is_quota_wall("max-lane", 0, WALL, "a quoted " + WALL)
    other = WALL.replace("429", "400")
    assert not is_quota_wall("max-lane", 1, other, None)


class Link:
    def __init__(self):
        self.flushed = 0

    def flush(self):
        self.flushed += 1


def test_a_drain_counts_its_jobs_and_flushes_once_nothing_is_left():
    results = iter([True, True, False])
    link = Link()
    ran = drain_on_demand(None, "node-a", link, "max", step=lambda *_: next(results))
    assert (ran, link.flushed) == (2, 1)


def test_a_failing_iteration_ends_the_drain_without_raising():
    def boom(*_):
        raise RuntimeError

    link = Link()
    assert drain_on_demand(None, "node-a", link, "max", step=boom) == 0
    assert link.flushed == 1
