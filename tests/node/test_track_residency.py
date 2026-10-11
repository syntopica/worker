from worker.node.node_memory import NodeMemory
from worker.node.track_residency import track_residency


def test_loads_and_unloads_are_counted_over_the_last_hour():
    memory = NodeMemory()
    assert track_residency(memory, ["m"], 0.0) == {"loads_1h": 0, "unloads_1h": 0}
    track_residency(memory, [], 10.0)
    assert track_residency(memory, ["m"], 20.0) == {"loads_1h": 1, "unloads_1h": 1}
    assert track_residency(memory, ["m"], 3615.0) == {"loads_1h": 1, "unloads_1h": 0}
    assert track_residency(memory, ["m"], 3700.0) == {"loads_1h": 0, "unloads_1h": 0}


def test_each_change_is_logged_with_its_time(capsys):
    memory = NodeMemory()
    track_residency(memory, ["m"], 0.0)
    assert capsys.readouterr().err == ""
    track_residency(memory, ["n"], 10.0)
    lines = capsys.readouterr().err.splitlines()
    assert [line.rsplit(" at=", 1)[0] for line in lines] == [
        "worker: residency load model=n",
        "worker: residency unload model=m",
    ]
    assert all(line.endswith("Z") for line in lines)
