import worker


def test_package_exposes_contract_version():
    assert worker.CONTRACT_VERSION == 1
