package jobs

import (
	"errors"
	"maps"
	"strings"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	schemaLocation         = "https://worker.invalid/job-schema"
	schemaRegexMatchBudget = 100 * time.Millisecond
)

// checkOutputSchema returns the schema rule that failed and whether the
// output was accepted. Validation messages can contain the output itself.
func checkOutputSchema(input map[string]any, output any) (string, bool) {
	body := object(output)
	if body == nil {
		return "", false
	}
	schema, found := input[keySchema]
	if !found {
		schema = input[keyOutputSchema]
	}
	if schema == nil {
		return "", true
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(schemaLoader{})
	patterns := schemaRegexes{remaining: schemaRegexMatchBudget}
	compiler.UseRegexpEngine(patterns.compile)
	if err := compiler.AddResource(schemaLocation, currentSchema(schema)); err != nil {
		return "", false
	}
	compiled, err := compiler.Compile(schemaLocation)
	if err != nil {
		return "", false
	}
	return validateOutputSchema(compiled, body["json"], &patterns)
}

func validateOutputSchema(schema *jsonschema.Schema, value any, patterns *schemaRegexes) (string, bool) {
	err := schema.Validate(value)
	// A match error is indeterminate, including under `not` or `if`.
	if patterns.failed {
		return "", false
	}

	if err != nil {
		return failingSchemaPath(err), false
	}

	return "", true
}

func failingSchemaPath(err error) string {
	var failure *jsonschema.ValidationError
	if !errors.As(err, &failure) {
		return ""
	}
	unit := failure.DetailedOutput()
	for len(unit.Errors) > 0 && (unit.KeywordLocation == "" || strings.HasSuffix(unit.KeywordLocation, "/$ref")) {
		unit = &unit.Errors[0]
	}
	return unit.KeywordLocation
}

// currentSchema keeps root validation on the contract's draft even when a
// producer includes a draft hint. Invalid non-string hints still fail schema
// validation and local references still resolve inside the supplied schema.
func currentSchema(schema any) any {
	fields := object(schema)
	if _, hinted := fields["$schema"].(string); !hinted {
		return schema
	}
	current := maps.Clone(fields)
	delete(current, "$schema")
	return current
}
