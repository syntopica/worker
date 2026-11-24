package api

import (
	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

type request struct {
	store     *store.Store
	principal auth.Principal
	cfg       *config.Config
	id        string
	raw       []byte
	query     query
	reveal    string
	now       float64
}
