import json
import time
from pathlib import Path

from tests.tasks.fake_runner import fake_runner
from worker.config.task_profile import TaskProfile
from worker.tasks.read_usage import read_usage
from worker.tasks.run_task import run_task

FIXTURES = Path(__file__).parent / "fixtures"
CURSOR = (FIXTURES / "cursor_envelope.json").read_text()
AGY = (FIXTURES / "agy_error_envelope.json").read_text()


def test_cursor_envelope_maps_input_and_output_tokens():
    assert read_usage("cursor", CURSOR) == {"tokens_in": 17584, "tokens_out": 54}


def test_agy_envelope_maps_its_snake_case_counts():
    assert read_usage("agy", AGY) == {"tokens_in": 0, "tokens_out": 0}
    envelope = json.loads(AGY)
    envelope["usage"].update(input_tokens=120, output_tokens=7)
    assert read_usage("agy", json.dumps(envelope)) == {"tokens_in": 120, "tokens_out": 7}


def test_codex_and_unreadable_envelopes_report_no_usage():
    assert read_usage("codex", "ok\n") == {}
    assert read_usage("cursor", "not json") == {}
    assert read_usage("cursor", json.dumps({"usage": {"inputTokens": "many"}})) == {}
    assert read_usage("agy", json.dumps({"usage": {"input_tokens": True}})) == {}
    assert read_usage("agy", "[1]") == {}


def test_run_task_reports_the_cursor_envelope_usage(tmp_path):
    prof = TaskProfile(
        name="p",
        runner="cursor",
        model=None,
        reasoning=None,
        privacy=frozenset({"internal"}),
        timeout_s=30.0,
        nodes=None,
        input_root=None,
        command=fake_runner(tmp_path, "cursor", f"print({CURSOR.strip()!r})"),
    )
    the_lease = {"attempt_id": "a", "generation": 1, "input": {"prompt": "p"}}
    report = run_task(prof, the_lease, "node-a", lambda: True, time.monotonic, time.sleep)
    assert report is not None
    assert report["outcome"] == "succeeded"
    assert report["usage"] == {"tokens_in": 17584, "tokens_out": 54}
