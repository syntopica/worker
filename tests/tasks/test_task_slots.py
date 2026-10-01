import threading

import pytest

from worker.tasks.run_task_slots import run_task_slots
from worker.tasks.stoppable_sleep import stoppable_sleep


def test_a_stopped_slot_sleep_raises_at_once():
    stop = threading.Event()
    sleep = stoppable_sleep(stop)
    sleep(0.01)
    stop.set()
    with pytest.raises(SystemExit):
        sleep(60)


def test_every_slot_runs_only_the_first_probes_and_all_stop_together():
    seen = []
    ended = []

    def run(config, node, link, *, sleep=None, current=None, probe=True):
        seen.append((link, probe))
        if sleep is None:
            raise SystemExit(0)
        try:
            while True:
                sleep(0.01)
        except SystemExit:
            ended.append(link)
            raise

    with pytest.raises(SystemExit):
        run_task_slots(None, "node-a", ["l0", "l1", "l2"], run=run)
    assert sorted(seen) == [("l0", True), ("l1", False), ("l2", False)]
    assert sorted(ended) == ["l1", "l2"]
