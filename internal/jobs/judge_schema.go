package jobs

const maxJudgeScore = 5

func judgeSchema(letters []string) map[string]any {
	properties := make(map[string]any, len(letters))
	for _, letter := range letters {
		properties[letter] = map[string]any{keyType: "integer", "minimum": 1, "maximum": maxJudgeScore}
	}
	return map[string]any{
		keyType: "object",
		"properties": map[string]any{
			"scores": map[string]any{
				keyType: "object", "properties": properties, "required": letters, "additionalProperties": false,
			},
			keyBest: map[string]any{keyType: "string", "enum": letters},
		},
		"required": []string{"scores", "best"}, "additionalProperties": false,
	}
}
