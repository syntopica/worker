package api

import "net/http"

func routes() []route {
	const producer, node, admin = "producer", "node", "admin"
	return []route{
		{http.MethodPost, "v1/jobs", producer, submit},
		{http.MethodGet, "v1/jobs/*", producer, getJob},
		{http.MethodPost, "v1/jobs/*/cancel", producer, cancelJob},
		{http.MethodPost, "v1/jobs/*/ack", producer, ack},
		{http.MethodGet, "v1/results", producer, results},
		{http.MethodPost, "v1/leases", node, lease},
		{http.MethodPost, "v1/attempts/*/heartbeat", node, heartbeat},
		{http.MethodPost, "v1/attempts/*/complete", node, complete},
		{http.MethodPost, "v1/nodes/*/report", node, nodeReport},
		{http.MethodPost, "v1/nodes/*/walls", node, runnerWalls},
		{http.MethodGet, "v1/status", admin, status},
		{http.MethodGet, "v1/activity", admin, activity},
		{http.MethodGet, "v1/costs", admin, costs},
		{http.MethodGet, "v1/quality", admin, quality},
		{http.MethodGet, "v1/admin/jobs", admin, adminJobs},
		{http.MethodGet, "v1/admin/jobs/*", admin, adminJob},
		{http.MethodGet, "v1/admin/jobs/*/content", admin, adminContent},
		{http.MethodPost, "v1/admin/jobs/*/cancel", admin, adminCancel},
		{http.MethodPost, "v1/admin/jobs/*/retry", admin, adminRetry},
		{http.MethodPost, "v1/admin/jobs/*/ack", admin, adminAck},
		{http.MethodGet, "v1/admin/audit", admin, adminAudit},
	}
}
