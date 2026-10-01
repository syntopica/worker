"""The cooldowns row a quota wall rests: one model of a runner, or the whole runner."""


def cooldown_key(runner: str, model: str | None) -> str:
    """``runner:model`` when the profile pins a model, else the runner alone.

    One runner meters its models separately (agy's default model can be spent
    while its Gemini models answer), so a wall rests only the model that hit
    it; a runner-wide reading, such as CodexBar's, rests the runner key.
    """
    return f"{runner}:{model}" if model else runner
