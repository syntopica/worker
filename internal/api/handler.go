package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

const requestTimeout = 45 * time.Second

// Handler serves contract v1 against a migrated store.
type Handler struct {
	store       *store.Store
	dir         string
	current     func() *config.Config
	routes      []route
	diagnostics io.Writer
}

// NewHandler creates a handler that takes one live configuration snapshot per request.
func NewHandler(s *store.Store, dir string, current func() *config.Config, diagnostics io.Writer) *Handler {
	return &Handler{store: s, dir: dir, current: current, routes: routes(), diagnostics: diagnostics}
}

// ServeHTTP authenticates the request and returns only allowlisted error codes.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	status, value, err := h.dispatch(ctx, w, r)
	if err != nil {
		h.replyError(w, err)
		return
	}
	if err := writeJSON(w, status, value); err != nil {
		h.replyError(w, err)
	}
}

func (h *Handler) replyError(w http.ResponseWriter, err error) {
	if refusal, ok := errors.AsType[*fault.Error](err); ok && refusal.Status != http.StatusInternalServerError {
		writeRefusal(w, refusal.Status, refusal.Code)
		return
	}
	_, _ = fmt.Fprintln(h.diagnostics, "worker: request failed")
	writeRefusal(w, http.StatusInternalServerError, "internal")
}

func (h *Handler) dispatch(ctx context.Context, w http.ResponseWriter, r *http.Request) (_ int, _ any, err error) {
	// The HTTP server's default panic log includes arbitrary panic values.
	defer func() {
		if recover() != nil {
			err = fault.New(http.StatusInternalServerError, "internal")
		}
	}()

	cfg := h.current()
	raw, err := readBody(w, r, cfg.MaxPayloadBytes)
	if err != nil {
		return 0, nil, err
	}
	principal, err := auth.Authenticate(h.dir, r.Header.Get("Authorization"))
	if err != nil {
		return 0, nil, err
	}
	match, id := findRoute(h.routes, r.Method, r.URL.EscapedPath())
	if match == nil {
		return 0, nil, fault.New(http.StatusNotFound, "no_route")
	}
	if principal.Kind != match.role {
		return 0, nil, forbidden()
	}
	return match.handle(ctx, &request{
		store: h.store, principal: principal, cfg: cfg, id: id,
		raw: raw, query: parseQuery(r.URL.RawQuery), reveal: r.Header.Get("X-Worker-Reveal"), now: epochSeconds(),
	})
}
