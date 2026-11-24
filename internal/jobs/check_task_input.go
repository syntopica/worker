package jobs

import (
	"slices"
	"strings"
)

const maxTaskInputs = 64

// checkTaskInput refuses with bad_input unless the input is a runnable
// read-only task (amendment 2026-09-30).
func checkTaskInput(input map[string]any) error {
	for name := range input {
		if !slices.Contains([]string{keyRunner, keyProfile, "prompt", keyInputs, "output_schema"}, name) {
			return bad(codeBadInput)
		}
	}
	profile, _ := input[keyProfile].(string)
	prompt, _ := input[keyPrompt].(string)
	_, runnerIsText := input[keyRunner].(string)
	_, hasRunner := input[keyRunner]
	_, schemaIsObject := input[keyOutputSchema].(map[string]any)
	_, hasSchema := input[keyOutputSchema]
	if profile == "" || prompt == "" || hasRunner && !runnerIsText || hasSchema && !schemaIsObject {
		return bad(codeBadInput)
	}
	return checkTaskInputs(input)
}

func checkTaskInputs(input map[string]any) error {
	inputs, isList := input[keyInputs].([]any)
	if _, present := input[keyInputs]; present && !isList || len(inputs) > maxTaskInputs {
		return bad(codeBadInput)
	}
	if !slices.ContainsFunc(inputs, unsafeRelativePath) {
		return nil
	}

	return bad(codeBadInput)
}

// unsafeRelativePath reports anything but a non-empty relative POSIX path
// with no ".." and no empty or "." parts.
func unsafeRelativePath(entry any) bool {
	path, _ := entry.(string)
	if path == "" || strings.ContainsAny(path, "\\\x00") {
		return true
	}

	return slices.ContainsFunc(strings.Split(path, "/"), func(part string) bool {
		return part == "" || part == "." || part == ".."
	})
}
