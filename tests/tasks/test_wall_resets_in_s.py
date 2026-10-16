from worker.tasks.wall_resets_in_s import wall_resets_in_s


def test_agy_weekly_wall_names_its_reset():
    text = "error: Individual quota reached. Please upgrade your subscription. Resets in 167h30m25s."
    assert wall_resets_in_s(text) == 167 * 3600 + 30 * 60 + 25


def test_minutes_only_and_absent():
    assert wall_resets_in_s("Resets in 45m") == 45 * 60
    assert wall_resets_in_s("Individual quota reached.") is None
