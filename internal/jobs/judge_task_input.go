package jobs

import (
	"crypto/sha256"
	"slices"
	"strings"
)

const rubric = "You are judging several answers to the same request, written by different systems." +
	" Score each answer from 1 (useless or wrong) to 5 (fully correct, complete and in the" +
	" requested format), judging only against the request. Then name the best answer."

func judgeTaskInput(group string, input map[string]any, rows []map[string]any) (map[string]any, error) {
	chat, err := inferenceTaskInput(input)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(rows, func(a, b map[string]any) int {
		first := sha256.Sum256([]byte(group + text(a[keyID])))
		second := sha256.Sum256([]byte(group + text(b[keyID])))
		return slices.Compare(first[:], second[:])
	})
	labels := make(map[string]any, len(rows))
	letters := make([]string, 0, len(rows))
	var shown []string
	for index, row := range rows {
		label := answerLabel(index)
		output, err := decodeValue([]byte(text(row[keyBody])))
		if err != nil {
			return nil, errInvalidOutput
		}
		executor, err := decodeValue([]byte(text(row[keyExecutor])))
		if err != nil {
			return nil, errInvalidResult
		}
		labels[label] = executor
		letters = append(letters, label)
		shown = append(shown, "=== Answer "+label+" ===\n"+text(object(output)["text"]))
	}
	prompt := rubric + "\n\n=== Request ===\n" + text(chat[keyPrompt]) + "\n\n" + strings.Join(shown, "\n\n")
	return map[string]any{keyPrompt: prompt, keyOutputSchema: judgeSchema(letters), keyLabels: labels}, nil
}

func answerLabel(index int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	position := index + 1
	var label string
	for position > 0 {
		offset := (position - 1) % len(alphabet)
		label = alphabet[offset:offset+1] + label
		position = (position - 1) / len(alphabet)
	}
	return label
}
