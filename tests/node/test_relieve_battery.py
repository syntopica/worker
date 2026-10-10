from worker.node import relieve_battery as relieve_battery_module
from worker.node.node_memory import NodeMemory
from worker.node.relieve_battery import relieve_battery


def _unloads(monkeypatch):
    unloaded = []
    monkeypatch.setattr(
        relieve_battery_module, "unload_model", lambda url, model: unloaded.append(model) or True
    )
    return unloaded


def test_battery_unloads_owned_models_once(monkeypatch):
    unloaded = _unloads(monkeypatch)
    memory = NodeMemory(owned={"model-a"})
    assert relieve_battery("u", memory, ["model-a", "other"], on_ac=False) == ["model-a"]
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == []
    assert unloaded == ["model-a"]


def test_ac_or_unknown_power_unloads_nothing(monkeypatch):
    unloaded = _unloads(monkeypatch)
    memory = NodeMemory(owned={"model-a"})
    assert relieve_battery("u", memory, ["model-a"], on_ac=True) == []
    assert relieve_battery("u", memory, ["model-a"], on_ac=None) == []
    assert unloaded == []


def test_a_model_never_owned_is_left_alone(monkeypatch):
    unloaded = _unloads(monkeypatch)
    assert relieve_battery("u", NodeMemory(), ["model-a"], on_ac=False) == []
    assert unloaded == []


def test_a_failed_drain_is_left_to_its_timed_retry(monkeypatch):
    unloaded = _unloads(monkeypatch)
    memory = NodeMemory(owned={"model-a"}, failed_model="model-a")
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == []
    assert unloaded == []


def test_a_failed_unload_is_not_retried_every_step(monkeypatch):
    calls = []
    monkeypatch.setattr(
        relieve_battery_module, "unload_model", lambda url, model: calls.append(model) and False
    )
    memory = NodeMemory(owned={"model-a"})
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == []
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == []
    assert calls == ["model-a"]


def test_a_failed_unload_keeps_ownership_and_is_retried_after_ac(monkeypatch):
    answers = iter([False, True])
    calls = []
    monkeypatch.setattr(
        relieve_battery_module,
        "unload_model",
        lambda url, model: calls.append(model) or next(answers),
    )
    memory = NodeMemory(owned={"model-a"})
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == []
    assert "model-a" in memory.owned
    assert relieve_battery("u", memory, ["model-a"], on_ac=True) == []
    assert relieve_battery("u", memory, ["model-a"], on_ac=False) == ["model-a"]
    assert calls == ["model-a", "model-a"]
