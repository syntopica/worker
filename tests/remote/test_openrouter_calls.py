import pytest

from tests.remote.fake_openrouter import FakeOpenRouter
from worker.remote.openrouter_headroom import openrouter_headroom
from worker.remote.openrouter_output import openrouter_output
from worker.remote.openrouter_request_body import openrouter_request_body
from worker.remote.post_openrouter import post_openrouter

SCHEMA = {"type": "object", "properties": {"label": {"type": "string"}}}


@pytest.fixture
def fake():
    server = FakeOpenRouter()
    yield server
    server.close()


def test_the_body_carries_a_strict_schema_and_zdr_when_asked():
    job = {
        "messages": [{"role": "user", "content": "hi"}],
        "schema": SCHEMA,
        "options": {"temperature": 0},
    }
    body = openrouter_request_body("vendor/m:free", job, zdr=True)
    assert body["response_format"]["json_schema"]["schema"] == SCHEMA
    assert body["provider"] == {"zdr": True}
    assert body["temperature"] == 0
    assert "provider" not in openrouter_request_body("vendor/m:free", job, zdr=False)


def test_output_parses_fenced_json_and_reports_cost():
    answer = {
        "choices": [{"message": {"content": '```json\n{"a": 1}\n```'}}],
        "usage": {"prompt_tokens": 5, "completion_tokens": 2, "cost": 0.25},
    }
    output, usage = openrouter_output(answer)
    assert output["json"] == {"a": 1}
    assert usage == {"tokens_in": 5, "tokens_out": 2, "cost_usd": 0.25}


def test_post_maps_answers_and_refusals_to_allowlisted_codes(fake):
    answer, error = post_openrouter("k", {"model": "m"}, 5, api=fake.api)
    assert error is None
    assert answer["choices"]
    fake.status = 429
    assert post_openrouter("k", {"model": "m"}, 5, api=fake.api) == (None, "rate_limited")
    fake.status, fake.answer = 200, {"error": {"message": "provider text"}}
    assert post_openrouter("k", {"model": "m"}, 5, api=fake.api) == (None, "bad_response")
    assert post_openrouter("k", {}, 1, api="http://127.0.0.1:1") == (None, "transport_error")


def test_headroom_reads_the_free_daily_counter_and_unknown_is_none(fake):
    assert openrouter_headroom("k", api=fake.api) == 5
    assert openrouter_headroom("k", api="http://127.0.0.1:1", timeout=1) is None
