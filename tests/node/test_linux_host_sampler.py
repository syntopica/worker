from worker.config.node_policy import NodePolicy
from worker.node.default_host_sampler import default_host_sampler
from worker.node.hold_power_source import HoldPowerSource
from worker.node.linux_host_sampler import LinuxHostSampler
from worker.node.parse_loadavg import parse_loadavg
from worker.node.parse_meminfo_free_pct import parse_meminfo_free_pct

MEMINFO = "MemTotal:       100000 kB\nMemFree:  1000 kB\nMemAvailable:    40000 kB\n"


def files(load: str | None, meminfo: str | None = MEMINFO):
    return lambda path: load if path == "/proc/loadavg" else meminfo


def test_parsers_read_the_kernel_formats_and_refuse_garbage():
    assert parse_loadavg("1.53 1.24 1.09 2/1803 4242\n") == 1.53
    assert parse_loadavg("") is None
    assert parse_meminfo_free_pct(MEMINFO) == 40.0
    assert parse_meminfo_free_pct("MemTotal: 100 kB\n") is None


def test_load_above_the_limit_resets_idle_like_a_returning_user():
    now = [100.0]
    load = ["2.0 0 0 1/1 1"]
    sampler = LinuxHostSampler(
        8.0, 20.0, lambda p: load[0] if "load" in p else MEMINFO, lambda: now[0]
    )
    now[0] = 400.0
    assert sampler().idle_s == 300.0
    load[0] = "9.5 0 0 1/1 1"
    now[0] = 410.0
    state = sampler()
    assert state.idle_s == 0.0
    assert state.on_ac is True
    assert state.pressure == "normal"
    load[0] = "3.0 0 0 1/1 1"
    now[0] = 470.0
    assert sampler().idle_s == 60.0


def test_without_a_limit_idle_counts_from_start():
    now = [0.0]
    sampler = LinuxHostSampler(None, 20.0, files("99.0 0 0 1/1 1"), lambda: now[0])
    now[0] = 50.0
    assert sampler().idle_s == 50.0


def test_low_memory_warns_and_unreadable_files_degrade():
    low = "MemTotal: 100 kB\nMemAvailable: 10 kB\n"
    assert LinuxHostSampler(None, 20.0, files("1.0", low))().pressure == "warn"
    state = LinuxHostSampler(None, 20.0, files(None, None))()
    assert state.idle_s is None
    assert state.pressure == "unknown"


def test_the_platform_picks_the_sampler():
    node = NodePolicy("n", "server", 300, 30, "http://127.0.0.1:11434", "", max_load=24.0)
    assert isinstance(default_host_sampler(node, "linux"), LinuxHostSampler)
    assert isinstance(default_host_sampler(node, "darwin"), HoldPowerSource)
