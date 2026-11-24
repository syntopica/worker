package jobs

import (
	"errors"
	"maps"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaLocation = "https://worker.invalid/job-schema"

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
	if err := compiler.AddResource(schemaLocation, currentSchema(schema)); err != nil {
		return "", false
	}
	compiled, err := compiler.Compile(schemaLocation)
	if err != nil {
		return "", false
	}
	err = compiled.Validate(body["json"])
	if err == nil {
		return "", true
	}
	return failingSchemaPath(err), false
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
