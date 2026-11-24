package jobs

import (
	"encoding/json"
	"slices"
	"unicode/utf8"
)

const (
	maxKeyLength    = 256
	maxPriority     = 100
	maxAttempts     = 10
	defaultPriority = "50"
	defaultAttempts = "3"
)

// parseSubmit validates a decoded submit body against contract v1 and names
// the first problem.
func parseSubmit(body map[string]any) (*submission, error) {
	sub, err := submitHeader(body)
	if err != nil {
		return nil, err
	}
	if boundsErr := submitBounds(body, sub); boundsErr != nil {
		return nil, boundsErr
	}
	input, isObject := body[keyInput].(map[string]any)
	if !isObject {
		return nil, bad("missing_input")
	}
	sub.Input = input
	if sub.Kind == kindTask {
		err = checkTaskInput(input)
	} else {
		err = checkInferenceInput(input)
	}
	if err != nil {
		return nil, err
	}

	return sub, submitTail(body, sub)
}

func submitHeader(body map[string]any) (*submission, error) {
	sub := &submission{Tier: "basic", Queue: text(body[keyQueue])}
	if tier, present := body[keyTier]; present {
		sub.Tier, _ = tier.(string)
	}
	sub.Kind, _ = body[keyKind].(string)
	sub.Privacy, _ = body[keyPrivacy].(string)
	sub.Key, _ = body["idempotency_key"].(string)
	if err := validateSubmitHeader(body, sub); err != nil {
		return nil, err
	}
	if sub.Kind == kindInference {
		models, _ := object(body["requirements"])["models"].([]any)
		if len(models) == 0 {
			return nil, bad("missing_models")
		}
		for _, model := range models {
			sub.Models = append(sub.Models, text(model))
		}
	}
	if sub.Key == "" || utf8.RuneCountInString(sub.Key) > maxKeyLength {
		return nil, bad("bad_idempotency_key")
	}
	return sub, nil
}

func validateSubmitHeader(body map[string]any, sub *submission) error {
	switch {
	case !isOne(body["contract"]):
		return bad("unsupported_contract")
	case sub.Kind != kindInference && sub.Kind != kindTask:
		return bad("unsupported_kind")
	case sub.Privacy != "public" && sub.Privacy != "internal" && !Sensitive(sub.Privacy):
		return bad("unknown_privacy")
	case sub.Tier != "basic" && sub.Tier != "strong":
		return bad("unknown_tier")
	}
	return nil
}

// isOne accepts what compares equal to the number one: 1, 1.0 and true.
func isOne(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case json.Number:
		number, err := typed.Float64()

		return err == nil && number == 1
	default:
		return false
	}
}

func submitBounds(body map[string]any, sub *submission) error {
	fields := map[string]any{keyPriority: json.Number(defaultPriority), keyMaxAttempts: json.Number(defaultAttempts)}
	for name := range fields {
		if value, present := body[name]; present {
			fields[name] = value
		}
	}
	priority, err := coerceNumber(fields[keyPriority], true)
	if err != nil {
		return err
	}
	attempts, err := coerceNumber(fields[keyMaxAttempts], true)
	if err != nil {
		return err
	}
	if priority < 0 || priority > maxPriority || attempts < 1 || attempts > maxAttempts {
		return bad(codeOutOfRange)
	}
	sub.Priority, sub.MaxAttempts = int(priority), int(attempts)

	return nil
}

func submitTail(body map[string]any, sub *submission) error {
	if body["parent_id"] != nil {
		parent, ok := body["parent_id"].(string)
		if !ok {
			return bad(codeBadSplitChild)
		}
		sub.ParentID = &parent
	}
	if body["deadline"] != nil {
		deadline, err := coerceNumber(body["deadline"], false)
		if err != nil {
			return err
		}
		sub.Deadline = &deadline
	}
	if sub.Kind != kindInference {
		return nil
	}
	return parseExclusions(object(body["requirements"]), sub)
}

func parseExclusions(requirements map[string]any, sub *submission) error {
	excluded, present := requirements["exclude_providers"]
	if !present {
		return nil
	}
	names, isList := excluded.([]any)
	if !isList {
		return bad("bad_requirements")
	}
	for _, name := range names {
		provider, _ := name.(string)
		if provider == "" {
			return bad("bad_requirements")
		}
		if !slices.Contains(sub.Exclude, provider) {
			sub.Exclude = append(sub.Exclude, provider)
		}
	}
	return nil
}
