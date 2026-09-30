import argparse
import threading

from worker.cli import cmd_serve as cmd_serve_module
from worker.cli.cmd_serve import cmd_serve


def test_interrupt_closes_the_server_and_returns_0(served, monkeypatch):
    events = []

    class FakeServer:
        def serve_forever(self):
            events.append("serve")
            raise KeyboardInterrupt

        def server_close(self):
            events.append("close")

    def fake_build(_config, _state, _current=None):
        events.append("build")
        return FakeServer()

    class FakeThread:
        def __init__(self, *_args, **_kwargs):
            pass

        def start(self):
            events.append("sweeper")

    monkeypatch.setattr(cmd_serve_module, "build_server", fake_build)
    real_migrate = cmd_serve_module.migrate_state
    monkeypatch.setattr(
        cmd_serve_module,
        "migrate_state",
        lambda state: events.append("migrate") or real_migrate(state),
    )
    monkeypatch.setattr(
        cmd_serve_module, "reconcile_lost_payloads", lambda _c, _now: events.append("reconcile")
    )
    monkeypatch.setattr(threading, "Thread", FakeThread)
    assert cmd_serve(argparse.Namespace()) == 0
    assert events == ["migrate", "reconcile", "build", "sweeper", "serve", "close"]
