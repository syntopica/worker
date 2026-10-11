import io
import re
import time

import pytest

from tests.node.fake_ollama import FakeOllama
from tests.node.test_run_attempt import IDLE, TYPING, Link, lease, nap, node, states
from worker.config.model_pin import ModelPin
from worker.node.ollama_call import OllamaCall
from worker.node.ollama_request_body import ollama_request_body
from worker.node.read_chat_stream import read_chat_stream
from worker.node.run_attempt import run_attempt

PIN = ModelPin("model-a", 40960, "5m", 30, 2)


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def read(raw):
    started = []
    answer, error = read_chat_stream(io.BytesIO(raw), lambda: started.append(True))
    return answer, error, len(started)


def test_chunks_are_joined_into_the_final_answer():
    raw = b'{"message":{"content":"a"},"done":false}\n\n{"message":{"content":"b"},"done":true,"eval_count":2}\n'
    answer, error, started = read(raw)
    assert error is None
    assert answer == {"message": {"content": "ab"}, "done": True, "eval_count": 2}
    assert started == 1


@pytest.mark.parametrize(
    "raw",
    [
        b'{"error":"x"}\n',
        b"not json\n",
        b"[1]\n",
        b'{"message":{"content":"a"},"done":false}\n',
        b"",
    ],
)
def test_a_bad_or_unfinished_stream_is_a_bad_response(raw):
    assert read(raw)[:2] == (None, "bad_response")


def wait_phase(fake, seconds):
    call = OllamaCall.start(fake.url, ollama_request_body(PIN, {"messages": []}), timeout=10)
    time.sleep(seconds)
    phase = call.phase()
    call.cancel()
    return phase


def test_the_phase_turns_to_generation_with_the_first_chunk(fake):
    fake.delay = 2.0
    assert wait_phase(fake, 0.3) == "prefill"
    fake.delay, fake.generation_delay = 0.0, 2.0
    assert wait_phase(fake, 0.5) == "generation"


def test_the_drain_line_names_phase_reason_and_time(fake, capsys):
    fake.generation_delay = 3.0
    outcome = run_attempt(
        lease(), Link(), PIN, node(fake.url), states(IDLE, TYPING), nap, time.time
    )
    assert outcome == "preempted"
    line = capsys.readouterr().err
    assert re.search(
        r"worker: drain quiet in \d+\.\ds phase=generation reason=user_active"
        r" at=\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ",
        line,
    )
