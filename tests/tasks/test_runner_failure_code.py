import pytest

from worker.tasks.runner_failure_code import runner_failure_code


@pytest.mark.parametrize(
    ("stderr", "code"),
    [
        ("ERROR: unexpected status 401 Unauthorized", "runner_auth"),
        ("error: 429 Too Many Requests", "rate_limited"),
        ("stream disconnected before completion", "runner_unavailable"),
        ("HTTP 503 from upstream", "runner_unavailable"),
        ("panic: something else", "runner_failed"),
        ("", "runner_failed"),
    ],
)
def test_the_stderr_tail_names_the_failure(stderr, code):
    assert runner_failure_code(stderr) == code


def test_only_the_tail_is_read():
    assert runner_failure_code("401 unauthorized" + "x" * 5000) == "runner_failed"
