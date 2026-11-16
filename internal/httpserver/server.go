package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Server struct { name, address string; logger *slog.Logger }

func New(name, address string, logger *slog.Logger) *Server {
	return &Server{name: name, address: address, logger: logger}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status":"ok", "service":s.name})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status":"ready", "service":s.name})
	})
	server := &http.Server{Addr:s.address, Handler:mux, ReadHeaderTimeout:5*time.Second}
	errCh := make(chan error, 1)
	go func() { s.logger.Info("http server listening", "service", s.name, "address", s.address); errCh <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil { return err }
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) { return nil }; return err
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) { return nil }; return err
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
