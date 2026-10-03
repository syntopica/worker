import dataclasses
import time

from tests.api.test_api_roundtrip import api, call  # noqa: F401
from worker.config.task_profile import TaskProfile
from worker.tasks.probe_runner_walls import probe_runner_walls
from worker.tasks.spent_until import spent_until

LATER = "2030-01-01T00:00:00Z"
LATER_TS = 1893456000.0


def test_a_spent_known_window_reports_its_reset_and_unknown_usage_does_not():
    usage = {"primary": {"usedPercent": 100, "resetsAt": LATER}, "secondary": {"usedPercent": 10}}
    assert spent_until(usage, (), 0.0) == LATER_TS
    unknown = {"primary": {"usedPercent": 100, "usageKnown": False, "resetsAt": LATER}}
    assert spent_until(unknown, (), 0.0) is None
    assert spent_until(usage, (), LATER_TS + 1) is None


def test_named_windows_read_only_the_matching_extras():
    usage = {
        "primary": {"usedPercent": 100, "resetsAt": LATER},
        "extraRateWindows": [
            {"id": "gemini-5h", "window": {"usedPercent": 20, "resetsAt": LATER}},
            {"id": "claude-5h", "window": {"usedPercent": 99, "resetsAt": LATER}},
        ],
    }
    assert spent_until(usage, ("gemini",), 0.0) is None
    assert spent_until(usage, ("claude",), 0.0) == LATER_TS


def test_the_probe_maps_each_configured_runner(config):
    wired = dataclasses.replace(
        config, runner_quota={"cursor": ("cursor", ()), "codex": ("codex", ())}
    )
    usages = {"cursor": {"primary": {"usedPercent": 100, "resetsAt": LATER}}, "codex": None}
    assert probe_runner_walls(wired, 0.0, read=usages.get) == {"cursor": LATER_TS}


def test_a_node_reported_wall_rests_the_runner(api):  # noqa: F811
    base, t = api
    until = time.time() + 3600
    status, _ = call(base, t["node-a"], "POST", "/v1/nodes/node-a/walls", {"cursor": until})
    assert status == 200
    status, _ = call(base, t["node-a"], "POST", "/v1/nodes/node-b/walls", {"cursor": until})
    assert status == 403
    status, _ = call(base, t["node-a"], "POST", "/v1/nodes/node-a/walls", {"cursor": until * 2})
    assert status == 400
    status, body = call(base, t["admin"], "GET", "/v1/status")
    assert round(body["cooldowns"]["cursor"]) in (3599, 3600)


def test_model_windows_rest_only_the_profiles_whose_model_the_spent_window_meters(config):
    profiles = {
        name: TaskProfile(name, "runner-a", model, None, frozenset(), 60.0, None, None)
        for name, model in (("p.gemini", "gemini-pro"), ("p.claude", "claude-sonnet"))
    }
    wired = dataclasses.replace(
        config,
        profiles=profiles,
        runner_quota={"runner-a": ("provider-a", ())},
        runner_model_windows={"runner-a": {"gemini": ("gemini",), "3p": ("claude", "gpt")}},
    )
    usage = {
        "primary": {"usedPercent": 100, "resetsAt": LATER},
        "extraRateWindows": [
            {"id": "quota-gemini-weekly", "window": {"usedPercent": 98, "resetsAt": LATER}},
            {"id": "quota-3p-weekly", "window": {"usedPercent": 0, "resetsAt": LATER}},
        ],
    }
    assert probe_runner_walls(wired, 0.0, read=lambda _: usage) == {"runner-a:gemini-pro": LATER_TS}
