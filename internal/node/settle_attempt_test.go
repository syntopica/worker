package node

import (
	"testing"
	"time"
)

func at(seconds int64) time.Time { return time.Unix(seconds, 0) }

func TestTheBackoffDoublesCapsAtTwoHoursAndResetsOnSuccess(t *testing.T) {
	t.Parallel()

	m := newMemory()

	for _, want := range []int64{900, 1800, 3600, 7200, 7200} {
		settleAttempt(m, testModel, true, preempted, memoryPressure, at(0))

		if !m.backoffUntil.Equal(at(want)) || m.pressureStreak != pressureSamples {
			t.Fatalf("backoff until %v, want %d", m.backoffUntil, want)
		}
	}

	settleAttempt(m, testModel, true, succeeded, "", at(10))
	settleAttempt(m, testModel, true, preempted, memoryPressure, at(10))

	if !m.backoffUntil.Equal(at(910)) {
		t.Fatalf("backoff until %v", m.backoffUntil)
	}
}

func TestConsecutiveTransportFailuresBackOffAndSuccessResets(t *testing.T) {
	t.Parallel()

	m := newMemory()
	m.normalSince = new(at(0))

	for _, want := range []int64{1030, 1060} {
		if !settleAttempt(m, testModel, true, failed, transportError, at(1000)) || !m.transportUntil.Equal(at(want)) {
			t.Fatalf("transport until %v, want %d", m.transportUntil, want)
		}
	}

	if nodeBlock(m, idleState(), at(1059), false) != transportBackoff || nodeBlock(m, idleState(), at(1060), false) != "" {
		t.Fatal("the transport backoff does not end at its time")
	}

	for range 20 {
		settleAttempt(m, testModel, true, failed, transportError, at(1000))
	}

	if !m.transportUntil.Equal(at(2800)) {
		t.Fatalf("transport until %v", m.transportUntil)
	}

	if settleAttempt(m, testModel, true, failed, "http_500", at(2000)) || m.transportFailures != 0 {
		t.Fatal("another failure counted as transport")
	}
}

func TestOnlyAnAttemptThatRanMakesTheModelTheNodesOwn(t *testing.T) {
	t.Parallel()

	m := newMemory()

	for _, why := range []reason{transportError, nodeError, "http_404"} {
		settleAttempt(m, testModel, false, failed, why, at(0))
	}

	if len(m.owned) != 0 {
		t.Fatal("a failed attempt owns the model")
	}

	settleAttempt(m, testModel, false, drainFailed, userActive, at(0))

	if !m.owned[testModel] || m.failedModel != testModel {
		t.Fatalf("memory: %+v", m)
	}
}
