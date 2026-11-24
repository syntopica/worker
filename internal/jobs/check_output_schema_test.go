package jobs

import "testing"

func TestOutputSchemaUsesTheContractDraft(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		schema map[string]any
		output any
		valid  bool
	}{
		"unknown dialect hint": {
			schema: map[string]any{"$schema": "https://example.invalid/dialect", "type": "string"},
			output: "answer", valid: true,
		},
		"older tuple items": {
			schema: map[string]any{
				"$schema": "http://json-schema.org/draft-07/schema#", "type": "array",
				"items": []any{map[string]any{"type": "string"}},
			},
			output: []any{"answer"},
		},
		"remote reference": {
			schema: map[string]any{"$ref": "https://example.invalid/schema"},
			output: "answer",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, valid := checkOutputSchema(map[string]any{"schema": tc.schema}, map[string]any{"json": tc.output})
			if valid != tc.valid {
				t.Fatalf("output schema accepted = %t; want %t", valid, tc.valid)
			}
		})
	}
}
