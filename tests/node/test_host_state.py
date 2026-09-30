from worker.node.admission_block import admission_block
from worker.node.host_state import HostState
from worker.node.parse_hid_idle_seconds import parse_hid_idle_seconds
from worker.node.parse_on_ac import parse_on_ac
from worker.node.parse_pressure import parse_pressure
from worker.node.release_reason import release_reason
from worker.node.sample_host_state import sample_host_state

IOREG = '    | |     "HIDIdleTime" = 16558114625\n'


def test_parsers():
    assert parse_hid_idle_seconds(IOREG) == 16.558114625
    assert parse_hid_idle_seconds("nothing") is None
    assert parse_on_ac("Now drawing from 'AC Power'\n") is True
    assert parse_on_ac("Now drawing from 'Battery Power'\n") is False
    assert (
        parse_pressure("1\n"),
        parse_pressure("2"),
        parse_pressure("4"),
        parse_pressure("x"),
    ) == ("normal", "warn", "critical", "unknown")


def test_a_failed_read_counts_as_busy():
    state = sample_host_state(lambda args: None)
    assert release_reason(state, "active_ok") == "host_state_unreadable"


def test_user_input_releases_idle_only_work_but_not_active_ok():
    typing = HostState(2.0, True, "normal")
    assert release_reason(typing, "idle") == "user_active"
    assert release_reason(typing, "active_ok") is None


def test_battery_and_pressure_release_everything():
    assert release_reason(HostState(900, False, "normal"), "active_ok") == "on_battery"
    assert release_reason(HostState(900, True, "warn"), "active_ok") == "memory_pressure"


def test_admission_waits_for_pressure_recovery():
    normal = HostState(900, True, "normal")
    assert admission_block(normal, normal_since=100.0, now=150.0) == "pressure_recovering"
    assert admission_block(normal, normal_since=100.0, now=230.0) is None


def test_a_failed_reader_is_logged_by_name_only(capsys):
    sample_host_state(lambda args: None)
    err = capsys.readouterr().err
    assert err == "worker: host readers failed: ioreg,pmset,pressure_level\n"


def test_an_unparsable_reading_is_logged_as_a_parse_failure(capsys):
    sample_host_state(lambda args: "garbage")
    assert "ioreg_parse" in capsys.readouterr().err


def test_a_transient_pmset_failure_is_retried_once():
    calls = []

    def flaky(args):
        calls.append(args[0])
        if args[0] == "/usr/bin/pmset" and calls.count("/usr/bin/pmset") == 1:
            return None
        return {"/usr/bin/pmset": "Now drawing from 'AC Power'\n"}.get(args[0])

    state = sample_host_state(flaky)
    assert state.on_ac is True
    assert calls.count("/usr/bin/pmset") == 2
