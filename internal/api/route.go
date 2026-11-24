package api

import (
	"context"
	"strings"
)

type route struct {
	method string
	path   string
	role   string
	handle func(context.Context, *request) (int, any, error)
}

func findRoute(routes []route, method, path string) (matched *route, id string) {
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	for i := range routes {
		if routes[i].method != method {
			continue
		}
		if value, ok := matchPath(routes[i].path, parts); ok {
			return &routes[i], value
		}
	}
	var absent *route
	return absent, ""
}
