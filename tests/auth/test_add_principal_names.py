import pytest

from worker.auth.add_principal import add_principal


@pytest.mark.parametrize("name", ["../x", "a/b", "", ".hidden", "Upper", "a" * 65, "a\n"])
def test_unsafe_names_are_rejected_before_anything_is_written(tmp_path, name):
    with pytest.raises(ValueError, match="invalid principal name"):
        add_principal(tmp_path / "state", "producer", name)
    assert not (tmp_path / "state").exists()


@pytest.mark.parametrize("name", ["pa", "node-a", "a.b_c-1", "a" * 64])
def test_ordinary_names_are_accepted(tmp_path, name):
    assert add_principal(tmp_path / "state", "producer", name)
