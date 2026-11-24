package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type nodeReportRow struct {
	Reason   string      `json:"reason"`
	Idle     json.Number `json:"idle_s"`
	Resident []string    `json:"resident"`
	Age      float64     `json:"age_s"`
	Loads    json.Number `json:"loads_1h"`
	Unloads  json.Number `json:"unloads_1h"`
	Release  *struct {
		Code string  `json:"code"`
		Age  float64 `json:"age_s"`
	} `json:"last_release"`
}

func renderNodes(raw json.RawMessage) (string, error) {
	var report struct {
		Nodes map[string]nodeReportRow `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return "", fmt.Errorf("read node status: %w", err)
	}

	var out strings.Builder
	for _, name := range slices.Sorted(maps.Keys(report.Nodes)) {
		node := report.Nodes[name]
		renderNode(&out, name, &node)
	}
	return out.String(), nil
}

func renderNode(out *strings.Builder, name string, node *nodeReportRow) {
	reason := node.Reason
	if reason == "" {
		reason = "working/ready"
	}
	fmt.Fprintf(out, "%-20s %-22s idle=%s resident=%s age=%.0fs",
		name, reason, reportNumber(node.Idle), renderResident(node.Resident), node.Age)
	if node.Release != nil {
		fmt.Fprintf(out, " last_release=%s@%.0fs", node.Release.Code, node.Release.Age)
	}
	if node.Loads != "" {
		fmt.Fprintf(out, " loads_1h=%s unloads_1h=%s", node.Loads, reportNumber(node.Unloads))
	}
	out.WriteByte('\n')
}

func renderResident(models []string) string {
	if models == nil {
		return "None"
	}
	quoted := make([]string, 0, len(models))
	for _, model := range models {
		escaped := strconv.Quote(model)
		if strings.Contains(model, "'") && !strings.Contains(model, `"`) {
			quoted = append(quoted, escaped)
			continue
		}
		unquoted := strings.ReplaceAll(escaped[1:len(escaped)-1], `\"`, `"`)
		quoted = append(quoted, "'"+strings.ReplaceAll(unquoted, "'", `\'`)+"'")
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
