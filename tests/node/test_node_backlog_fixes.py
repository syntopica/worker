import time

import pytest

from tests.node.fake_ollama import FakeOllama
from worker.config.model_pin import ModelPin
from worker.config.node_policy import NodePolicy
from worker.node.host_state import HostState
from worker.node.node_block import node_block
from worker.node.node_memory import NodeMemory
from worker.node.ollama_call import OllamaCall
from worker.node.retry_failed_drain import retry_failed_drain
from worker.node.run_attempt import run_attempt
from worker.node.settle_attempt import settle_attempt

PIN = ModelPin("model-a", 40960, "5m", 30, 2)
IDLE = HostState(900, True, "normal")
LEASE = {
    "job_id": "j",
    "attempt_id": "a",
    "generation": 1,
    "model": "model-a",
    "input": {"messages": []},
    "run_when": "idle",
}


class BrokenLink:
    def heartbeat(self, attempt_id, generation, draining):
        raise RuntimeError("link broke")

    def complete(self, attempt_id, generation, report):
        pass


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def test_consecutive_transport_failures_back_off_and_success_resets():
    memory = NodeMemory(normal_since=0.0)
    assert settle_attempt(memory, "m", True, "failed", "transport_error", 1000.0)
    assert memory.transport_until == 1030.0
    settle_attempt(memory, "m", True, "failed", "transport_error", 1000.0)
    assert memory.transport_until == 1060.0
    assert node_block(memory, IDLE, 1059.0, False) == "transport_backoff"
    assert node_block(memory, IDLE, 1060.0, False) is None
    for _ in range(20):
        settle_attempt(memory, "m", True, "failed", "transport_error", 1000.0)
    assert memory.transport_until == 1000.0 + 1800.0
    settle_attempt(memory, "m", True, "succeeded", None, 2000.0)
    assert memory.transport_failures == 0


def test_other_failures_do_not_start_a_transport_backoff():
    memory = NodeMemory()
    assert not settle_attempt(memory, "m", True, "failed", "http_500", 1000.0)
    assert memory.transport_until == 0.0


def test_a_failed_drain_is_retried_every_five_minutes_while_listed():
    memory = NodeMemory(failed_model="model-a")
    sent = []

    def unload(url, model):
        sent.append(model)
        return False

    assert retry_failed_drain("u", memory, ["model-a"], 100.0, unload)
    assert not retry_failed_drain("u", memory, ["model-a"], 399.0, unload)
    assert retry_failed_drain("u", memory, ["model-a"], 400.0, unload)
    assert not retry_failed_drain("u", memory, [], 900.0, unload)
    assert sent == ["model-a", "model-a"]


def test_an_exception_mid_attempt_cancels_the_backend_call(fake, monkeypatch):
    fake.delay = 5.0
    cancelled = []
    original = OllamaCall.cancel

    def spy(self):
        cancelled.append(True)
        original(self)

    monkeypatch.setattr(OllamaCall, "cancel", spy)
    node = NodePolicy("node-a", "owner", 300, 44, fake.url, "label-a")
    with pytest.raises(RuntimeError):
        run_attempt(LEASE, BrokenLink(), PIN, node, lambda: IDLE, lambda _s: None, time.time)
    assert cancelled == [True]
