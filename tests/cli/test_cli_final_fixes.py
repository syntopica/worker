import json
import plistlib
import socket
import urllib.error
import urllib.request
from pathlib import Path

import pytest

from tests.cli.test_cli import instance
from tests.conftest import CONFIG
from worker.api import make_handler as make_handler_module
from worker.auth.add_principal import add_principal
from worker.cli.main import main


def test_an_unknown_node_name_is_one_line_and_exit_2(tmp_path, monkeypatch, capsys):
    state = instance(tmp_path, monkeypatch)
    add_principal(state, "node", "node-zz")  # a token alone must not start the loop
    assert main(["node", "--name", "node-zz"]) == 2
    err = capsys.readouterr().err
    assert err.count("\n") == 1
    assert "Traceback" not in err


def test_a_port_in_use_is_one_line_and_exit_2(tmp_path, monkeypatch, capsys):
    holder = socket.socket()
    holder.bind(("127.0.0.1", 0))
    holder.listen()
    try:
        instance(tmp_path, monkeypatch)
        cfg = {**CONFIG, "listen": f"127.0.0.1:{holder.getsockname()[1]}"}
        (tmp_path / "worker" / "config.json").write_text(json.dumps(cfg))
        assert main(["serve"]) == 2
    finally:
        holder.close()
    err = capsys.readouterr().err
    assert err.count("\n") == 1
    assert "in use" in err


def test_a_500_logs_the_class_and_route_without_the_payload(served, monkeypatch, capsys):
    def explode(_ctx):
        raise RuntimeError("payload text")

    monkeypatch.setattr(make_handler_module, "authenticate", lambda *_args: None)
    monkeypatch.setattr(make_handler_module, "route_request", explode)
    port = json.loads((served / "worker" / "config.json").read_text())["listen"].split(":")[1]
    with pytest.raises(urllib.error.HTTPError) as error:
        urllib.request.urlopen(f"http://127.0.0.1:{port}/v1/status?secret=x", timeout=5)
    assert error.value.code == 500
    err = capsys.readouterr().err
    assert "worker: 500 RuntimeError GET /v1/status" in err
    assert "payload text" not in err
    assert "secret" not in err


@pytest.mark.parametrize("agent", ["serve", "node"])
def test_both_launch_agents_throttle_restarts(agent):
    template = (
        Path(__file__).parents[2] / "launchd" / f"com.syntopica.worker.{agent}.plist.template"
    )
    plist = plistlib.loads(template.read_text().encode())
    assert plist["ThrottleInterval"] == 60
