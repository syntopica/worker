import time

import pytest

from tests.node.fake_ollama import FakeOllama
from worker.config.model_pin import ModelPin
from worker.node.chat_output import chat_output
from worker.node.ollama_call import OllamaCall
from worker.node.ollama_request_body import ollama_request_body
from worker.node.probe_quiet import probe_quiet
from worker.node.resident_models import resident_models
from worker.node.restart_ollama import restart_ollama
from worker.node.unload_model import unload_model

PIN = ModelPin("model-a", 40960, "5m", 30, 2)


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def test_pinned_options_cannot_be_overridden_by_the_producer():
    body = ollama_request_body(
        PIN,
        {
            "messages": [],
            "options": {"num_ctx": 2048, "temperature": 0},
            "schema": {"type": "object"},
        },
    )
    assert body["options"] == {"temperature": 0, "num_ctx": 40960}
    assert (body["keep_alive"], body["think"], body["stream"], body["format"]) == (
        "5m",
        False,
        False,
        {"type": "object"},
    )


def test_call_returns_parsed_output_and_usage(fake):
    call = OllamaCall.start(fake.url, ollama_request_body(PIN, {"messages": []}), timeout=10)
    while not call.done():
        time.sleep(0.01)
    answer, error = call.outcome()
    assert error is None
    assert chat_output(answer) == (
        {"text": '{"label": "x"}', "json": {"label": "x"}},
        {"tokens_in": 7, "tokens_out": 3},
    )


def test_cancel_ends_the_call_with_a_cancelled_code(fake):
    fake.delay = 5.0
    call = OllamaCall.start(fake.url, ollama_request_body(PIN, {"messages": []}), timeout=10)
    time.sleep(0.2)
    call.cancel()
    deadline = time.time() + 2
    while not call.done() and time.time() < deadline:
        time.sleep(0.01)
    assert call.outcome() == (None, "cancelled")


def test_probe_reports_quiet_only_after_the_backend_frees(fake):
    fake.busy_until = time.time() + 1.5
    assert probe_quiet(fake.url, PIN, timeout=0.5) is False
    assert probe_quiet(fake.url, PIN, timeout=3.0) is True


def test_resident_and_unload(fake):
    assert resident_models(fake.url) == ["model-a"]
    assert unload_model(fake.url, "model-a") is True
    assert resident_models(fake.url) == []


def test_restart_uses_the_users_gui_domain(monkeypatch):
    seen = []
    monkeypatch.setattr(
        "worker.node.restart_ollama.run_command", lambda args, timeout=5.0: seen.append(args) or ""
    )
    assert restart_ollama("svc.label") is True
    assert seen[0][:3] == ["/bin/launchctl", "kickstart", "-k"] and seen[0][3].endswith(
        "/svc.label"
    )
    monkeypatch.setattr("worker.node.restart_ollama.run_command", lambda args, timeout=5.0: None)
    assert restart_ollama("svc.label") is False
