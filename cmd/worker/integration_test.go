package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type plistElement struct {
	plistValue

	kind string
}

type plistValue struct {
	Text  string         `xml:",chardata"`
	Items []plistElement `xml:",any"`
}

func (element *plistElement) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	element.kind = start.Name.Local
	return decoder.DecodeElement(&element.plistValue, &start)
}

func TestLaunchdTemplates(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		agent, processType string
		runnerPath         bool
	}{
		{agent: "serve"},
		{agent: "node", processType: "Background"},
		{agent: "tasks", processType: "Standard", runnerPath: true},
		{agent: "remote", processType: "Background", runnerPath: true},
	} {
		t.Run(tc.agent, func(t *testing.T) {
			t.Parallel()
			const binary, data, name, path = "/opt/example/bin/worker", "/var/example/instance", "example-node", "/opt/example/runners:/usr/bin"
			label := "com.syntopica.worker." + tc.agent
			rendered := renderLaunchAgent(t, label, binary, data, name, path)
			want := []string{binary, tc.agent}
			if tc.agent != "serve" {
				want = append(want, "--name", name)
			}
			arguments := rendered["ProgramArguments"]
			if got := plistArguments(t, arguments); !slices.Equal(got, want) {
				t.Errorf("ProgramArguments = %q, want %q", got, want)
			}
			assertLaunchAgent(t, rendered, label, data, tc.agent, tc.processType)
			environment := plistDictionary(t, rendered["EnvironmentVariables"])
			if got := environment["SYNTOPICA_DATA"].Text; got != data {
				t.Errorf("SYNTOPICA_DATA = %q, want %q", got, data)
			}
			value, exists := environment["PATH"]
			if exists != tc.runnerPath || (exists && value.Text != path) {
				t.Errorf("PATH present = %t, value = %q, want present = %t", exists, value.Text, tc.runnerPath)
			}
		})
	}
}

func renderLaunchAgent(t *testing.T, label, binary, data, name, path string) map[string]plistElement {
	t.Helper()
	template, err := os.ReadFile(filepath.Join("..", "..", "launchd", label+".plist.template"))
	if err != nil {
		t.Fatal(err)
	}
	rendered := strings.NewReplacer(
		"@WORKER_GO_BIN@", binary, "@SYNTOPICA_DATA@", data, "@NODE@", name, "@PATH@", path,
	).Replace(string(template))
	if strings.Contains(rendered, "@") {
		t.Fatal("rendered plist still contains a placeholder")
	}
	var plist plistElement
	if err := xml.Unmarshal([]byte(rendered), &plist); err != nil {
		t.Fatal(err)
	}
	if plist.kind != "plist" || len(plist.Items) != 1 {
		t.Fatal("expected one plist dictionary")
	}
	return plistDictionary(t, plist.Items[0])
}

func plistDictionary(t *testing.T, element plistElement) map[string]plistElement {
	t.Helper()
	if element.kind != "dict" || len(element.Items)%2 != 0 {
		t.Fatal("invalid plist dictionary")
	}
	entries := make(map[string]plistElement, len(element.Items)/2)
	for i := 0; i < len(element.Items); i += 2 {
		key := element.Items[i]
		if key.kind != "key" {
			t.Fatal("plist dictionary entry has no key")
		}
		if _, exists := entries[key.Text]; exists {
			t.Fatalf("duplicate plist key %q", key.Text)
		}
		entries[key.Text] = element.Items[i+1]
	}
	return entries
}

func plistArguments(t *testing.T, element plistElement) []string {
	t.Helper()
	if element.kind != "array" {
		t.Fatal("ProgramArguments must be an array")
	}
	values := make([]string, 0, len(element.Items))
	for _, item := range element.Items {
		if item.kind != "string" {
			t.Fatal("ProgramArguments must contain strings")
		}
		values = append(values, item.Text)
	}
	return values
}

func assertLaunchAgent(t *testing.T, entries map[string]plistElement, label, data, agent, processType string) {
	t.Helper()
	for key, want := range map[string]string{
		"Label":             label,
		"StandardErrorPath": data + "/worker/state/" + agent + ".err.log",
		"ProcessType":       processType,
		"ThrottleInterval":  "60",
	} {
		if got := entries[key].Text; got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if _, exists := entries["ProcessType"]; exists && processType == "" {
		t.Error("coordinator should use launchd's default process type")
	}
	for _, key := range []string{"RunAtLoad", "KeepAlive"} {
		if got := entries[key].kind; got != "true" {
			t.Errorf("%s = %q, want true", key, got)
		}
	}
}
