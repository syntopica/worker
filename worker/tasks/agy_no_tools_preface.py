"""The line agy's prompt opens with, so the model answers instead of reaching for a tool."""

# agy has no flag that removes its tools. In print mode a tool request is
# denied, and the turn still ends SUCCESS with an empty response: on
# 2026-10-14 a 195 KB prompt drew an empty answer in 2 of 4 runs without
# this line and in none of 3 with it.
AGY_NO_TOOLS_PREFACE = (
    "Answer in plain text from this prompt alone. Do not use any tool, run any"
    " command or read any file: everything you need is below.\n\n"
)
