import pytest

from tests.conftest import body
from tests.jobs.test_complete_attempt import report, run
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.known_error_code import known_error_code
from worker.jobs.submit_job import submit_job


@pytest.mark.parametrize(
    ("code", "kept"),
    [
        (None, None),
        ("runner_failed", "runner_failed"),
        ("http_503", "http_503"),
        ("unknown_profile", "unknown_profile"),
        ("http_5030", "executor_error"),
        ("input missing at /home/someone/notes.md", "executor_error"),
        ("made_up_code", "executor_error"),
    ],
)
def test_only_known_codes_pass(code, kept):
    assert known_error_code(code) == kept


def test_a_completion_with_free_text_settles_as_an_executor_error(conn, config, capsys):
    submit_job(conn, config, "pa", body(), 0.0)
    lease = run(conn, config)
    text = "Traceback: secret prompt text"
    complete_attempt(
        conn, config, lease.attempt_id, lease.generation, report("failed", code=text), 2.0
    )
    stored = conn.execute("SELECT error FROM attempts").fetchone()["error"]
    assert stored == "executor_error"
    assert "secret" not in capsys.readouterr().err
