from worker.client.worker_client import WorkerClient


def test_results_socket_timeout_outlives_the_requested_wait(monkeypatch):
    seen = {}

    def fake_call(self, method, path, payload=None, timeout=40.0):
        seen["timeout"] = timeout
        return {"results": []}

    monkeypatch.setattr(WorkerClient, "_call", fake_call)
    client = WorkerClient("http://127.0.0.1:1", "token")
    client.results("q", wait=120)
    assert seen["timeout"] > 120
    client.results("q")
    assert seen["timeout"] >= 10
