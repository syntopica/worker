import json
import time
import urllib.error
import urllib.request

import pytest

from worker.auth.add_principal import add_principal
from worker.store.open_store import open_store


def test_a_read_request_is_served_while_the_writer_lock_is_held(served):
    state = served / "worker" / "state"
    token = add_principal(state, "producer", "pa")
    port = json.loads((served / "worker" / "config.json").read_text())["listen"].split(":")[1]
    writer = open_store(state)
    writer.execute("BEGIN IMMEDIATE")
    try:
        started = time.monotonic()
        request = urllib.request.Request(
            f"http://127.0.0.1:{port}/v1/jobs/missing", headers={"Authorization": f"Bearer {token}"}
        )
        with pytest.raises(urllib.error.HTTPError) as error:
            urllib.request.urlopen(request, timeout=5)
        assert error.value.code == 404
        assert time.monotonic() - started < 2.0
    finally:
        writer.execute("ROLLBACK")
