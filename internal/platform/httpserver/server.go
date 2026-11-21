package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	name   string
	server *http.Server
	logger *slog.Logger
}

func New(name, address string, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		name: name, logger: logger,
		server: &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second},
	}
}

func (s *Server) Run() error {
	requests, cancelRequests := context.WithCancel(context.Background())
	defer cancelRequests()
	s.server.BaseContext = func(net.Listener) context.Context { return requests }
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("http server listening", "service", s.name, "address", s.server.Addr)
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-shutdownSignal():
		cancelRequests()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}
