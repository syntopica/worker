from tests.node.test_coordinator_link import link, serve  # noqa: F401


def test_an_undelivered_completion_is_held_and_blocks_new_leases(link):  # noqa: F811
    link.complete("a", 1, {"outcome": "succeeded"})
    assert link.pending == [("a", 1, {"outcome": "succeeded"})]
    assert link.lease_task(False, 900.0) is None
    assert len(link.pending) == 1


def test_a_held_completion_is_delivered_before_the_next_lease():
    server, spooled = serve(409)
    try:
        spooled.pending.append(("a", 1, {"outcome": "succeeded"}))
        spooled.lease_remote()
        assert spooled.pending == []
    finally:
        server.shutdown()
