package cli

import "encoding/json"

type reportLayout struct {
	path     string
	synopsis string
	timed    bool
	rows     bool
	render   func(json.RawMessage) (string, error)
}

func reportSpec(name string) reportLayout {
	const statusPath = "/v1/status"
	const timedSynopsis = "[--days DAYS] [--json]"
	switch name {
	case "nodes":
		return reportLayout{path: statusPath, synopsis: "[--json]", render: renderNodes}
	case "costs":
		return reportLayout{path: "/v1/costs", synopsis: timedSynopsis, timed: true, rows: true, render: renderCosts}
	case "quality":
		return reportLayout{path: "/v1/quality", synopsis: timedSynopsis, timed: true, render: renderQuality}
	case "audit":
		return reportLayout{path: "/v1/admin/audit", synopsis: timedSynopsis, timed: true, rows: true, render: renderAudit}
	default:
		return reportLayout{path: statusPath, synopsis: "[--json]", render: renderStatus}
	}
}
