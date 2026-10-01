from worker.cli.format_executor import format_executor


def test_names_provider_and_model():
    executor = {"node": "node-a", "provider": "openrouter", "model": "org/model:free"}
    assert format_executor(executor) == "openrouter org/model:free"


def test_marks_what_the_node_did_not_report():
    assert format_executor({"node": "node-a"}) == "unknown unknown"
    assert format_executor(None) == "unknown unknown"
