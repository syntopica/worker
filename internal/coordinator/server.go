package coordinator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/syntopica/worker/internal/api"
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/jobs"
	"github.com/syntopica/worker/internal/store"
)

const (
	startTimeout    = 30 * time.Second
	socketTimeout   = 30 * time.Second
	shutdownTimeout = 5 * time.Second
)

// Server owns a listening loopback socket and the coordinator's store.
type Server struct {
	listener net.Listener
	http     *http.Server
	store    *store.Store
	source   *config.Source
	log      io.Writer
}

// Listen prepares the instance database and binds the configured loopback listener.
// Pass the returned server to Run to serve requests and release these resources.
func Listen(ctx context.Context, configPath, stateDir string, diagnostics io.Writer) (*Server, error) {
	ctx, cancel := context.WithTimeout(ctx, startTimeout)
	defer cancel()
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	output := &diagnosticWriter{dest: diagnostics}
	source := config.NewSource(configPath, cfg, output)
	address := net.JoinHostPort(cfg.ListenHost, strconv.Itoa(cfg.ListenPort))
	var binder net.ListenConfig
	listener, err := binder.Listen(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("bind coordinator: %w", err)
	}
	s, err := prepareStore(ctx, stateDir)
	if err != nil {
		return nil, errors.Join(err, listener.Close())
	}
	httpServer := &http.Server{
		Handler:           api.NewHandler(s, stateDir, source.Current, output),
		ReadHeaderTimeout: socketTimeout,
		ReadTimeout:       socketTimeout,
		IdleTimeout:       socketTimeout,
		ErrorLog:          log.New(io.Discard, "", 0),
	}
	return &Server{listener: listener, http: httpServer, store: s, source: source, log: output}, nil
}

// Addr is the bound address, including the assigned port when configured with port zero.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// Run serves until ctx is cancelled or the listener fails, then releases its resources.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Requests keep their own deadline during shutdown, so a node can finish
	// delivering a completion while the coordinator drains.
	shutdown := make(chan error, 1)
	context.AfterFunc(ctx, func() { shutdown <- s.shutdown(ctx) })

	swept := make(chan struct{})
	go func() {
		defer close(swept)
		s.runSweeper(ctx)
	}()

	serveErr := s.http.Serve(s.listener)
	cancel()
	shutdownErr := <-shutdown
	<-swept
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, shutdownErr, s.store.Close())
}

func (s *Server) shutdown(ctx context.Context) error {
	grace, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(grace); err != nil {
		return errors.Join(err, s.http.Close())
	}
	return nil
}

func prepareStore(ctx context.Context, stateDir string) (*store.Store, error) {
	if err := store.Migrate(ctx, stateDir); err != nil {
		return nil, err
	}
	s, err := store.Open(ctx, stateDir)
	if err != nil {
		return nil, err
	}
	if _, err := jobs.ReconcileLostPayloads(ctx, s, epochSeconds()); err != nil {
		return nil, errors.Join(err, s.Close())
	}
	return s, nil
}
