package jobs

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

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

func TestOutputSchemaSearchesPatterns(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		schema map[string]any
		output any
		valid  bool
		path   string
	}{
		"positive lookahead searches inside the value": {
			schema: map[string]any{"pattern": `item(?=:)`},
			output: "next item: ready", valid: true,
		},
		"positive lookahead rejects a nonmatch": {
			schema: map[string]any{"pattern": `item(?=:)`},
			output: "next item is ready",
			path:   "/pattern",
		},
		"negative lookahead filters property names": {
			schema: map[string]any{
				"patternProperties":    map[string]any{`^item(?!_draft)`: map[string]any{"type": "string"}},
				"additionalProperties": false,
			},
			output: map[string]any{"item1": "ready"}, valid: true,
		},
		"negative lookahead excludes property names": {
			schema: map[string]any{
				"patternProperties":    map[string]any{`^item(?!_draft)`: map[string]any{"type": "string"}},
				"additionalProperties": false,
			},
			output: map[string]any{"item_draft": "ready"},
			path:   "/additionalProperties",
		},
		"backreference matches the same group": {
			schema: map[string]any{"pattern": `^(a|b)\1$`},
			output: "bb", valid: true,
		},
		"Unicode digits match": {
			schema: map[string]any{"pattern": `^\d$`},
			output: "٣", valid: true,
		},
		"dollar matches before a final newline": {
			schema: map[string]any{"pattern": `^item$`},
			output: "item\n", valid: true,
		},
		"standard pattern still rejects a nonmatch": {
			schema: map[string]any{"pattern": `^item-[0-9]+$`},
			output: "item-ready",
			path:   "/pattern",
		},
		"invalid pattern is rejected": {
			schema: map[string]any{"pattern": `(`},
			output: "(",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, valid := checkOutputSchema(map[string]any{"schema": tc.schema}, map[string]any{"json": tc.output})
			if valid != tc.valid {
				t.Fatalf("output schema accepted = %t; want %t", valid, tc.valid)
			}
			if path != tc.path {
				t.Fatalf("output schema path = %q; want %q", path, tc.path)
			}
		})
	}
}

func TestOutputSchemaRejectsAnIndeterminateNegatedPattern(t *testing.T) {
	t.Parallel()
	patterns := schemaRegexes{remaining: schemaRegexMatchBudget}
	compiler := jsonschema.NewCompiler()
	compiler.UseRegexpEngine(patterns.compile)
	schema := map[string]any{"not": map[string]any{"pattern": `item(?=:)`}}
	if err := compiler.AddResource(schemaLocation, schema); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(schemaLocation)
	if err != nil {
		t.Fatal(err)
	}
	patterns.remaining = 0
	if path, valid := validateOutputSchema(compiled, "next item is ready", &patterns); valid || path != "" {
		t.Fatalf("output schema accepted = %t, path = %q after match budget was exhausted", valid, path)
	}
}

func TestOutputSchemaSearchesPythonNamedPatterns(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		pattern string
		value   string
		valid   bool
	}{
		"named capture and lookahead use Unicode digits":             {`(?P<token>\d+)(?=:)`, "code ٣:", true},
		"named backreference preserves the final newline rule":       {`(?P<token>\w+):(?P=token)$`, "été:été\n", true},
		"named backreference rejects a different value":              {`(?P<token>item):(?P=token)$`, "item:other", false},
		"escaped group punctuation stays literal":                    {`\(\?P<literal>\)(?P<token>item)`, "(?P<literal>)item", true},
		"group punctuation in a class stays literal":                 {`(?P<token>item)[](?P<x>]+`, "item]P", true},
		"inline comment does not change the named reference":         {`(?P<token>item)(?# [ comment)(?P=token)`, "itemitem", true},
		"verbose comment does not hide the next group":               {"(?x)# [ comment\n(?P<token>item)(?P=token)", "itemitem", true},
		"escaped closing parenthesis stays inside an inline comment": {`(?P<token>item)(?# escaped \) (?P<token> ignored)(?P=token)`, "itemitem", true},
		"escaped newline stays inside a verbose comment":             {"(?x)# hidden \\\n(?P<token> ignored\n(?P<token>item)(?P=token)", "itemitem", true},
		"unclosed inline comment is invalid":                         {`(?P<token>item)(?# unterminated \)`, "item", false},
		"escaped opening parenthesis is not a declaration":           {`(?P<token>item)\(?P<literal>\)`, "item(P<literal>)", true},
		"duplicate named capture is invalid":                         {`(?P<token>item)(?P<token>item)`, "itemitem", false},
		"reference to a later group is invalid":                      {`(?P=token)?(?P<token>item)`, "item", false},
		"reference to an open group is invalid":                      {`(?P<token>(?P=token)?item)`, "item", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			schema := map[string]any{"pattern": tc.pattern}
			_, valid := checkOutputSchema(map[string]any{"schema": schema}, map[string]any{"json": tc.value})
			if valid != tc.valid {
				t.Fatalf("output schema accepted = %t; want %t", valid, tc.valid)
			}
		})
	}
}
