package client

import "net/url"

// jobPath percent-encodes the id, "/" included, so it stays one path segment.
func jobPath(jobID string) string {
	return "/v1/jobs/" + url.PathEscape(jobID)
}
