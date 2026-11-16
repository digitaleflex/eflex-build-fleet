package controlplane

import (
	"fmt"
	"net/http"
	"sync"
)

type metrics struct {
	mu       sync.Mutex
	counters map[string]int64
}

func newMetrics() *metrics {
	return &metrics{counters: map[string]int64{}}
}

func (m *metrics) inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

func (m *metrics) snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		out[k] = v
	}
	return out
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	snap := s.metrics.snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for name, v := range snap {
		fmt.Fprintf(w, "%s %d\n", name, v)
	}
}