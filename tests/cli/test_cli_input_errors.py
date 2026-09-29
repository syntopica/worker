import urllib.request

import pytest


def test_missing_job_file_is_exit_2_naming_the_file(served, run):
    run("token", "add", "--kind", "producer", "--name", "pa")
    code, out, err = run("submit", "--producer", "pa", str(served / "absent.json"))
    assert (code, out) == (2, "")
    assert err == f"worker: FileNotFoundError: {served / 'absent.json'}\n"


def test_bad_json_names_the_class_and_never_the_content(served, run):
    run("token", "add", "--kind", "producer", "--name", "pa")
    (served / "bad.json").write_text("secret-content {")
    code, _, err = run("submit", "--producer", "pa", str(served / "bad.json"))
    assert (code, err) == (2, "worker: JSONDecodeError\n")


def test_bad_config_is_exit_2_without_a_traceback(served, run):
    (served / "worker" / "config.json").write_text("not json at all")
    code, _, err = run("status")
    assert code == 2
    assert err.startswith("worker: ")
    assert "not json" not in err
    assert "Traceback" not in err


@pytest.mark.parametrize("error", [ConnectionResetError, TimeoutError])
def test_transport_failures_are_unreachable_exit_1(served, run, monkeypatch, error):
    run("token", "add", "--kind", "producer", "--name", "pa")

    def refuse(*_args, **_kwargs):
        raise error("secret-detail")

    monkeypatch.setattr(urllib.request, "urlopen", refuse)
    code, _, err = run("jobs", "--producer", "pa", "--queue", "pa.bulk")
    assert (code, err) == (1, f"worker: coordinator unreachable ({error.__name__})\n")
