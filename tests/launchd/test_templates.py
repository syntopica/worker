"""Test rendering and parsing of LaunchAgent plist templates."""

import plistlib
from pathlib import Path


def test_serve_template_renders_and_parses() -> None:
    """Test serve template rendering and validation."""
    # Read template
    template_path = (
        Path(__file__).parent.parent.parent
        / "launchd"
        / "com.syntopica.worker.serve.plist.template"
    )
    template_content = template_path.read_text()

    # Render with sample values
    rendered = template_content.replace("@WORKER_BIN@", "/usr/local/bin/worker")
    rendered = rendered.replace("@SYNTOPICA_DATA@", "/Users/test/Library/Syntopica")

    # Parse
    plist = plistlib.loads(rendered.encode("utf-8"))

    # Assert Label
    assert plist["Label"] == "com.syntopica.worker.serve"

    # Assert ProgramArguments
    assert plist["ProgramArguments"] == ["/usr/local/bin/worker", "serve"]

    # Assert EnvironmentVariables
    assert plist["EnvironmentVariables"]["SYNTOPICA_DATA"] == "/Users/test/Library/Syntopica"

    # Assert RunAtLoad and KeepAlive
    assert plist["RunAtLoad"] is True
    assert plist["KeepAlive"] is True

    # Assert stderr path
    assert plist["StandardErrorPath"].endswith("serve.err.log")

    # Assert no placeholder remains
    assert "@" not in rendered


def test_node_template_renders_and_parses() -> None:
    """Test node template rendering and validation."""
    # Read template
    template_path = (
        Path(__file__).parent.parent.parent / "launchd" / "com.syntopica.worker.node.plist.template"
    )
    template_content = template_path.read_text()

    # Render with sample values
    rendered = template_content.replace("@WORKER_BIN@", "/usr/local/bin/worker")
    rendered = rendered.replace("@SYNTOPICA_DATA@", "/Users/test/Library/Syntopica")
    rendered = rendered.replace("@NODE@", "node-a")

    # Parse
    plist = plistlib.loads(rendered.encode("utf-8"))

    # Assert Label
    assert plist["Label"] == "com.syntopica.worker.node"

    # Assert ProgramArguments
    assert plist["ProgramArguments"] == ["/usr/local/bin/worker", "node", "--name", "node-a"]

    # Assert EnvironmentVariables
    assert plist["EnvironmentVariables"]["SYNTOPICA_DATA"] == "/Users/test/Library/Syntopica"

    # Assert RunAtLoad and KeepAlive
    assert plist["RunAtLoad"] is True
    assert plist["KeepAlive"] is True

    # Assert ProcessType is Background
    assert plist["ProcessType"] == "Background"

    # Assert stderr path
    assert plist["StandardErrorPath"].endswith("node.err.log")

    # Assert no placeholder remains
    assert "@" not in rendered


def test_serve_template_no_process_type() -> None:
    """Test that serve template does not have ProcessType."""
    # Read template
    template_path = (
        Path(__file__).parent.parent.parent
        / "launchd"
        / "com.syntopica.worker.serve.plist.template"
    )
    template_content = template_path.read_text()

    # Render with sample values
    rendered = template_content.replace("@WORKER_BIN@", "/usr/local/bin/worker")
    rendered = rendered.replace("@SYNTOPICA_DATA@", "/Users/test/Library/Syntopica")

    # Parse
    plist = plistlib.loads(rendered.encode("utf-8"))

    # Assert ProcessType is not present
    assert "ProcessType" not in plist
