import pytest

from worker.config.parse_shadow_policy import parse_shadow_policy


def shadow(target):
    return {"rate": 0.1, "targets": [target], "judge": "judge.codex"}


def test_a_shadow_may_target_the_free_router():
    policy = parse_shadow_policy("q", shadow("openrouter:openrouter/free"))
    assert policy is not None
    assert policy.targets == ("openrouter:openrouter/free",)


def test_a_shadow_may_not_target_a_paid_router():
    with pytest.raises(ValueError, match=":free"):
        parse_shadow_policy("q", shadow("openrouter:openrouter/auto"))
