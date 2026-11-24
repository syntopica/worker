package cli

import (
	"io"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/link"
)

type executor struct {
	name, token string
	cfg         *config.Config
	policy      config.NodePolicy
	current     func() *config.Config
	diagnostics io.Writer
	http        *http.Client
}

func loadExecutor(name string, diagnostics io.Writer) (*executor, error) {
	path, state, err := resolvePaths()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	policy, ok := cfg.Nodes[name]
	if !ok {
		return nil, userError("node is not in the configuration")
	}
	token, err := readToken(state, name)
	if err != nil {
		return nil, err
	}
	log := &serialLog{out: diagnostics}
	source := config.NewSource(path, cfg, log)
	return &executor{name: name, token: token, cfg: cfg, policy: policy, current: source.Current, diagnostics: log, http: http.DefaultClient}, nil
}

func (e *executor) link(url string) *link.Coordinator {
	return link.New(url, e.token, e.name, e.http)
}
