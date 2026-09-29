import json
import threading

import pytest

from tests.conftest import body
from worker.api.build_server import build_server
from worker.auth.add_principal import add_principal
from worker.client.api_failure import ApiFailure
from worker.client.worker_client import WorkerClient


@pytest.fixture
def client(config, tmp_path):
    state = tmp_path / "state"
    token = add_principal(state, "producer", "pa")
    server = build_server(config, state)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    yield WorkerClient(f"http://127.0.0.1:{server.server_address[1]}", token)
    server.shutdown()


def test_submit_is_idempotent_and_conflicts_raise(client):
    job = json.loads(body())
    first = client.submit(job)
    assert client.submit(job)["id"] == first["id"]
    with pytest.raises(ApiFailure) as error:
        client.submit({**job, "priority": 99})
    assert error.value.code == "idempotency_conflict"


def test_wait_for_times_out_with_none(client):
    job_id = client.submit(json.loads(body()))["id"]
    assert client.wait_for(job_id, timeout=0.2, poll=0.1) is None
