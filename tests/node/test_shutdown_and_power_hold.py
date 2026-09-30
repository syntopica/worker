import os
import signal
import threading

import pytest

from tests.conftest import body, fresh_store
from tests.node.test_run_attempt import IDLE, PIN, Link, fake, lease, node  # noqa: F401
from tests.remote.test_remote_step import LEASE as REMOTE_LEASE
from tests.remote.test_remote_step import Link as RemoteLink
from tests.tasks.fake_runner import fake_runner
from tests.tasks.test_task_step import LEASE as TASK_LEASE
from tests.tasks.test_task_step import Link as TaskLink
from tests.tasks.test_task_step import make_config
from worker.jobs.complete_attempt import complete_attempt
from worker.jobs.completion_report import CompletionReport
from worker.jobs.lease_job import lease_job
from worker.jobs.lease_request import LeaseRequest
from worker.jobs.submit_job import submit_job
from worker.node.hold_power_source import HoldPowerSource
from worker.node.host_state import HostState
from worker.node.raise_on_sigterm import raise_on_sigterm
from worker.node.run_attempt import run_attempt
from worker.remote import run_remote_attempt as attempt_module
from worker.tasks.task_step import task_step


def test_a_failed_power_read_reuses_a_recent_one_only():
    samples = [HostState(5, True, "normal"), HostState(5, None, "normal")]
    now = [0.0]
    hold = HoldPowerSource(
        lambda: samples.pop(0) if len(samples) > 1 else samples[0], lambda: now[0]
    )
    assert hold().on_ac is True
    now[0] = 100.0
    assert hold().on_ac is True
    now[0] = 400.0
    assert hold().on_ac is None


def test_without_any_reading_nothing_is_invented():
    hold = HoldPowerSource(lambda: HostState(5, None, "normal"), lambda: 0.0)
    assert hold().on_ac is None


def stop():
    raise SystemExit(0)


def test_a_stopped_node_hands_its_inference_job_back(fake):  # noqa: F811
    link = Link()
    fake.delay = 5.0
    with pytest.raises(SystemExit):
        run_attempt(lease(), link, PIN, node(fake.url), stop, lambda _s: None, lambda: 0.0)
    report = link.completed[0]
    assert (report["outcome"], report["error_code"]) == ("preempted", "node_shutdown")


def test_a_stopped_task_loop_kills_the_runner_and_hands_the_task_back(tmp_path):
    marker = tmp_path / "alive"
    body_src = f"time.sleep(0.5)\nopen({str(marker)!r}, 'w').write('x')\n"
    config = make_config(tmp_path, fake_runner(tmp_path, "codex", body_src))
    link = TaskLink(TASK_LEASE)
    with pytest.raises(SystemExit):
        task_step(config, "node-a", link, lambda: IDLE, lambda: 0.0, lambda _s: stop())
    assert link.completed[0][2]["error_code"] == "node_shutdown"
    threading.Event().wait(1.0)
    assert not marker.exists()


def test_a_stopped_remote_loop_hands_its_job_back(config, monkeypatch):
    monkeypatch.setattr(attempt_module, "_BEAT_S", 0.01)
    link = RemoteLink(REMOTE_LEASE)
    link.heartbeat = lambda *_a: stop()

    def hang(*_a):
        threading.Event().wait(5)
        return None, "node_error"

    with pytest.raises(SystemExit):
        attempt_module.run_remote_attempt(
            REMOTE_LEASE, link, "node-a", "k", config, lambda: 0.0, post=hang
        )
    assert link.completed[0]["error_code"] == "node_shutdown"


def test_a_shutdown_charges_neither_an_attempt_nor_a_preemption(config, tmp_path):
    conn = fresh_store(tmp_path / "state")
    job_id, _ = submit_job(conn, config, "pa", body(), 0.0)
    leased = lease_job(conn, config, LeaseRequest("node-a", None, False, 40.0, 9999.0), 1.0)
    executor = {"node": "node-a", "provider": "ollama", "model": "model-a"}
    report = CompletionReport("preempted", None, {}, executor, "node_shutdown", 1.0)
    assert complete_attempt(conn, config, leased.attempt_id, leased.generation, report, 2.0) == (
        "queued"
    )
    row = conn.execute("SELECT attempts, preemptions, error FROM jobs WHERE id=?", (job_id,))
    assert tuple(row.fetchone()) == (0, 0, "node_shutdown")


def test_sigterm_becomes_system_exit():
    previous = signal.getsignal(signal.SIGTERM)
    try:
        raise_on_sigterm()
        with pytest.raises(SystemExit):
            os.kill(os.getpid(), signal.SIGTERM)
            threading.Event().wait(1.0)
    finally:
        signal.signal(signal.SIGTERM, previous)
