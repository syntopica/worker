from worker.jobs.candidate import Candidate
from worker.jobs.lease_request import LeaseRequest
from worker.policy.candidate_eligible import candidate_eligible
from worker.policy.pick_job import pick_job
from worker.policy.privacy_allows import privacy_allows


def cand(
    job_id,
    *,
    model="model-a",
    queue="pa.bulk",
    priority=50,
    created=0.0,
    privacy="mail",
    parked=False,
    need=0.0,
):
    return Candidate(job_id, queue, model, priority, privacy, created, parked, need)


def req(resident=None, active=False, free=44.0, idle=900.0, node="node-a"):
    return LeaseRequest(node, resident, active, free, idle)


def test_mail_never_reaches_a_guest_node(config):
    assert privacy_allows(config, "mail", "ollama", "owner")
    assert not privacy_allows(config, "mail", "ollama", "guest")
    assert not privacy_allows(config, "mail", "openrouter", "owner")


def test_user_activity_admits_only_active_ok_queues(config):
    assert not candidate_eligible(cand("a"), req(active=True), config)
    assert candidate_eligible(cand("b", queue="pa.live"), req(active=True), config)


def test_resident_model_needs_warm_memory_only(config):
    assert candidate_eligible(cand("a"), req(resident="model-a", free=3.0), config)
    # Another resident model is evicted (OLLAMA_MAX_LOADED_MODELS=1), so the
    # cold footprint is checked against the node's whole budget.
    assert candidate_eligible(cand("a"), req(resident="model-b", free=3.0), config)
    assert not candidate_eligible(cand("a"), req(node="node-g"), config)


def test_parked_job_waits_for_a_long_enough_current_idle(config):
    parked = cand("a", parked=True, need=1200.0)
    assert not candidate_eligible(parked, req(idle=900.0), config)
    assert candidate_eligible(parked, req(idle=1300.0), config)


def test_resident_model_wins_until_another_model_is_too_old(config):
    jobs = [cand("b", model="model-b", priority=90, created=1000.0), cand("a", created=1000.0)]
    assert pick_job(jobs, req(resident="model-a"), config, {}, 1100.0).job_id == "a"
    assert pick_job(jobs, req(resident="model-a"), config, {}, 1000.0 + 3601).job_id == "b"


def test_priority_then_weighted_share_then_age(config):
    jobs = [cand("old", created=1.0), cand("live", queue="pa.live", created=5.0)]
    # equal priority: pa.live weight 3 with 3 recent runs (1.0) vs pa.bulk weight 1 with 2 (2.0)
    assert pick_job(jobs, req(), config, {"pa.bulk": 2, "pa.live": 3}, 10.0).job_id == "live"
    jobs.append(cand("urgent", priority=99, created=9.0))
    assert pick_job(jobs, req(), config, {}, 10.0).job_id == "urgent"
