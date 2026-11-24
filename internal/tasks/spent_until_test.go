package tasks

import (
	"testing"
	"time"
)

const (
	later   = "2030-01-01T00:00:00Z"
	laterTS = 1893456000.0
)

// window is a rate window used this much, resetting later.
func window(used float64) map[string]any {
	return map[string]any{"usedPercent": used, "resetsAt": later}
}

func TestASpentKnownWindowReportsItsReset(t *testing.T) {
	t.Parallel()

	epoch := time.Unix(0, 0)
	usage := map[string]any{"primary": window(100), "secondary": map[string]any{"usedPercent": 10.0}}

	if until, spent := spentUntil(usage, nil, epoch); !spent || until != laterTS {
		t.Errorf("spent window: %v %v", until, spent)
	}

	if _, spent := spentUntil(usage, nil, time.Unix(int64(laterTS)+1, 0)); spent {
		t.Error("a reset already past is no wall")
	}

	unknown := map[string]any{"primary": map[string]any{"usedPercent": 100.0, "usageKnown": false, "resetsAt": later}}
	if _, spent := spentUntil(unknown, nil, epoch); spent {
		t.Error("unknown usage is no evidence")
	}

	unreadable := map[string]any{"primary": map[string]any{"usedPercent": 100.0, "resetsAt": "soon"}, "tertiary": 3}
	if _, spent := spentUntil(unreadable, nil, epoch); spent {
		t.Error("a reset that is no time is skipped")
	}
}

func TestNamedWindowsReadOnlyTheMatchingExtras(t *testing.T) {
	t.Parallel()

	usage := map[string]any{
		"primary": window(100),
		"extraRateWindows": []any{
			map[string]any{"id": "gemini-5h", "window": window(20)},
			map[string]any{"title": "Claude 5h", "window": window(99)},
			"noise",
		},
	}

	if _, spent := spentUntil(usage, []string{"gemini"}, time.Unix(0, 0)); spent {
		t.Error("the gemini window is not spent")
	}

	if until, spent := spentUntil(usage, []string{"claude"}, time.Unix(0, 0)); !spent || until != laterTS {
		t.Errorf("claude window: %v %v", until, spent)
	}
}
