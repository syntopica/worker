package jobs

import (
	"encoding/json"
	"strconv"
)

func retryBody(job *jobRow, input map[string]any) (map[string]any, error) {
	excluded := []any{}
	if job.Exclude != "" {
		value, err := decodeValue([]byte(job.Exclude))
		items, isList := value.([]any)
		if err != nil || !isList {
			return nil, errInvalidExclusions
		}
		excluded = items
	}
	return map[string]any{
		"contract": json.Number("1"), keyKind: job.Kind, keyQueue: job.Queue, "idempotency_key": "retry:" + job.ID,
		keyPriority: json.Number(strconv.Itoa(job.Priority)), keyPrivacy: job.Privacy, keyTier: job.Tier,
		keyMaxAttempts: json.Number(strconv.Itoa(job.MaxAttempts)),
		"requirements": map[string]any{"models": []any{job.Model}, "exclude_providers": excluded},
		keyInput:       input,
	}, nil
}
