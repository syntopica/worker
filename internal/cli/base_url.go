package cli

import (
	"net"
	"strconv"
	"strings"

	"github.com/syntopica/worker/internal/config"
)

func baseURL(cfg *config.Config) string {
	host := strings.Trim(cfg.ListenHost, "[]")
	switch host {
	case "0.0.0.0":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(cfg.ListenPort))
}
