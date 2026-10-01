"""The judge task for one shadow group: request, lettered answers and a score schema."""

from typing import Any

_RUBRIC = (
    "You are judging several answers to the same request, written by different systems."
    " Score each answer from 1 (useless or wrong) to 5 (fully correct, complete and in the"
    " requested format), judging only against the request. Then name the best answer."
)


def judge_task_input(messages: list[dict[str, Any]], answers: dict[str, str]) -> dict[str, Any]:
    """``answers`` maps a letter to an answer's text; letters are the only labels shown."""
    request = "\n\n".join(f"[{m['role']}]\n{m['content']}" for m in messages)
    shown = "\n\n".join(f"=== Answer {letter} ===\n{text}" for letter, text in answers.items())
    letters = list(answers)
    schema = {
        "type": "object",
        "properties": {
            "scores": {
                "type": "object",
                "properties": {
                    letter: {"type": "integer", "minimum": 1, "maximum": 5} for letter in letters
                },
                "required": letters,
                "additionalProperties": False,
            },
            "best": {"type": "string", "enum": letters},
        },
        "required": ["scores", "best"],
        "additionalProperties": False,
    }
    prompt = f"{_RUBRIC}\n\n=== Request ===\n{request}\n\n{shown}"
    return {"prompt": prompt, "output_schema": schema}
