import pytest

from tests.conftest import body
from worker.cli.run_sweeper import run_sweeper
from worker.jobs.submit_job import submit_job
from worker.store.open_store import open_store


def test_one_pass_reclaims_an_expired_lease(config, tmp_path):
    state = tmp_path / "state"
    conn = open_store(state)
    job_id, _ = submit_job(conn, config, "pa", body(), 100.0)
    conn.execute("UPDATE jobs SET state='leased', lease_expires=200 WHERE id=?", (job_id,))
    conn.close()
    run_sweeper(state, config, forever=False, clock=lambda: 1000.0)
    conn = open_store(state)
    row = conn.execute("SELECT state, attempts FROM jobs WHERE id=?", (job_id,)).fetchone()
    conn.close()
    assert (row["state"], row["attempts"]) == ("queued", 1)


def test_a_failing_pass_is_logged_by_class_and_does_not_stop_the_loop(config, tmp_path, capsys):
    calls = []

    def sleep(_seconds):
        calls.append(1)
        if len(calls) == 2:
            raise KeyboardInterrupt

    blocker = tmp_path / "state"
    blocker.write_text("a file where the state directory should be")
    with pytest.raises(KeyboardInterrupt):
        run_sweeper(blocker, config, sleep=sleep)
    err = capsys.readouterr().err
    assert len(calls) == 2
    assert err.count("worker: sweeper pass failed: ") == 2
    assert "a file where" not in err
