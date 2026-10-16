from worker.jobs.trusted_wall_reset import trusted_wall_reset
from worker.tasks.wall_resets_in_s import wall_resets_in_s

WALL = "error: Individual quota reached. Please upgrade your subscription to increase your limits."


def test_agy_weekly_wall_names_its_reset():
    assert wall_resets_in_s("agy", f"{WALL} Resets in 167h30m25s.") == 167 * 3600 + 30 * 60 + 25


def test_a_duration_away_from_the_wall_sentence_is_ignored():
    prose = "The model wrote: Resets in 300h. " + "x" * 50
    assert wall_resets_in_s("agy", prose + WALL) is None
    assert wall_resets_in_s("agy", "Resets in 45m") is None


def test_the_coordinator_trusts_only_a_bounded_duration():
    assert trusted_wall_reset({"wall_resets_in_s": 7200}) == 7200.0
    for bad in (float("inf"), float("nan"), -5, 0, 10**9, True, "3600"):
        assert trusted_wall_reset({"wall_resets_in_s": bad}) is None
