package cli

import (
	"net/http"
	"time"

	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/config"
)

func (*command) client(principal string) (*client.Client, error) {
	const requestTimeout = 40 * time.Second
	path, state, err := resolvePaths()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	token, err := readToken(state, principal)
	if err != nil {
		return nil, err
	}
	return client.New(baseURL(cfg), token, &http.Client{Timeout: requestTimeout}), nil
}
