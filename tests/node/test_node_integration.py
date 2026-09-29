import dataclasses
import threading
import time

import pytest

from tests.conftest import body
from tests.node.fake_ollama import FakeOllama
from worker.api.build_server import build_server
from worker.auth.add_principal import add_principal
from worker.jobs.list_results import list_results
from worker.jobs.submit_job import submit_job
from worker.node.coordinator_link import CoordinatorLink
from worker.node.host_state import HostState
from worker.node.run_node import run_node
from worker.store.open_store import open_store


class StopLoopError(Exception):
    pass


@pytest.fixture
def fake():
    server = FakeOllama()
    yield server
    server.close()


def test_node_runs_a_submitted_job_to_a_listable_result(config, tmp_path, fake):
    state = tmp_path / "state"
    node_token = add_principal(state, "node", "node-a")
    add_principal(state, "producer", "pa")
    local = dataclasses.replace(config.nodes["node-a"], ollama_url=fake.url)
    config = dataclasses.replace(config, nodes={**config.nodes, "node-a": local})
    server = build_server(config, state)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    try:
        conn = open_store(state)
        submit_job(conn, config, "pa", body(privacy="public"), time.time())
        conn.close()
        link = CoordinatorLink(f"http://127.0.0.1:{server.server_address[1]}", node_token, "node-a")
        offset, rests = [0.0], []

        def sleep(seconds):
            if seconds < 10:
                time.sleep(0.05)
                return
            rests.append(seconds)
            offset[0] += 200.0
            if len(rests) == 2:
                raise StopLoopError

        with pytest.raises(StopLoopError):
            run_node(
                config,
                "node-a",
                link,
                sample=lambda: HostState(900, True, "normal"),
                sleep=sleep,
                clock=lambda: time.time() + offset[0],
            )
        conn = open_store(state)
        results = list_results(conn, "pa", "pa.bulk", 0, 10)
        conn.close()
        assert [(r["output"]["json"], r["executor"]["node"]) for r in results] == [
            ({"label": "x"}, "node-a")
        ]
    finally:
        server.shutdown()
        server.server_close()
