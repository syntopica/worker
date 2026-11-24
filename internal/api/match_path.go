package api

import "strings"

func matchPath(pattern string, parts []string) (string, bool) {
	segments := strings.Split(pattern, "/")
	if len(segments) != len(parts) {
		return "", false
	}
	var id string
	for i, segment := range segments {
		if segment == "*" {
			id = parts[i]
		} else if segment != parts[i] {
			return "", false
		}
	}
	return id, true
}
